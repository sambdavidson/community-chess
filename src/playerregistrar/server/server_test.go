package server

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/sambdavidson/community-chess/src/lib/auth/grpcplayertokens"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	pb "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

func TestServerE2E(t *testing.T) {
	lis, stop := startServer(t)
	defer stop()
	cli, close := createClient(t, lis)
	defer close()

	playerCount := 1000
	r := rand.New(rand.NewSource(1))
	players := make([]*testPlayer, playerCount)
	for i := 0; i < playerCount; i++ {
		players[i] = &testPlayer{
			username: fmt.Sprintf("player_%d", i),
		}
	}
	ts := newTestServer(r, cli)
	for i := 0; i < (playerCount * 20); i++ {
		p := players[int64(math.Floor(r.Float64()*float64(playerCount)))]
		ts.randomAction(t, p)
	}
}

/* HELPERS */

type testPlayer struct {
	id        string
	suffix    int32
	username  string
	token     string
	added     bool
	got       bool
	refreshed bool
}

// Returns a listener device and a stopfunc
func startServer(t *testing.T) (*bufconn.Listener, func()) {
	svr, err := New(&Opts{})
	if err != nil {
		t.Fatal(err)
	}
	lis := bufconn.Listen(1024 * 1024)
	cli, close := createClient(t, lis) // Self referencial client for interceptor.
	tokenIngress := grpcplayertokens.NewPlayerAuthIngress(grpcplayertokens.PlayerAuthIngressArgs{
		AutoRefreshCadence:     time.Minute,
		PlayersRegistrarClient: cli,
	})
	s := grpc.NewServer(grpc.UnaryInterceptor(tokenIngress.GetUnaryServerInterceptor(grpcplayertokens.Ignore)))
	pb.RegisterPlayersRegistrarServer(s, svr)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	return lis, func() {
		close()
		s.GracefulStop()
	}
}

func createClient(t *testing.T, lis *bufconn.Listener) (pb.PlayersRegistrarClient, func() error) {
	ctx := context.Background()
	bufDialer := func(string, time.Duration) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.DialContext(ctx, "testfoo", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	cli := pb.NewPlayersRegistrarClient(conn)
	return cli, conn.Close
}

func newTestServer(r *rand.Rand, cli pb.PlayersRegistrarClient) *testServer {
	ts := &testServer{
		r:   r,
		cli: cli,
	}
	ts.actions = []action{ts.register, ts.get, ts.login, ts.refresh}
	return ts
}

type action func(*testing.T, *testPlayer)

type testServer struct {
	r       *rand.Rand
	cli     pb.PlayersRegistrarClient
	actions []action
}

func (ts *testServer) register(t *testing.T, p *testPlayer) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	got, err := ts.cli.RegisterPlayer(ctx, &pb.RegisterPlayerRequest{
		Username: p.username,
	})
	if err != nil {
		t.Error(err)
		return
	}
	np := got.GetPlayer()
	if np == nil {
		t.Errorf("created player '%s' returned is <nil>", p.username)
	}
	comparePlayers(t, p, np)
	if !p.added {
		p.added = true
		p.id = np.GetId()
		p.suffix = np.GetNumberSuffix()
	} else {
		// lets just let that other player die off.
	}

}
func (ts *testServer) get(t *testing.T, p *testPlayer) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	got, err := ts.cli.GetPlayer(ctx, &pb.GetPlayerRequest{
		PlayerId: p.id,
	})
	if !p.added {
		if err == nil {
			t.Error("no error on getting non-existant player")
		}
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	comparePlayers(t, p, got.GetPlayer())
	if p.suffix != got.GetPlayer().GetNumberSuffix() {
		t.Error("players have mismatched suffix")
	}
}
func (ts *testServer) login(t *testing.T, p *testPlayer) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	got, err := ts.cli.Login(ctx, &pb.LoginRequest{
		Username:     p.username,
		NumberSuffix: p.suffix,
	})
	if !p.added {
		if err == nil {
			t.Error("no error on getting non-existant player")
		}
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	token := got.GetToken()
	if len(token) == 0 {
		t.Error("login token empty")
		return
	}
	p.token = token
}
func (ts *testServer) refresh(t *testing.T, p *testPlayer) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx, err := grpcplayertokens.AppendPlayerAuthToOutgoingContext(ctx, p.token)
	if err != nil {
		t.Error(err)
		return
	}
	md, _ := metadata.FromOutgoingContext(ctx)
	ctx = metadata.NewIncomingContext(ctx, md)
	got, err := ts.cli.RefreshToken(ctx, &pb.RefreshTokenRequest{})
	if !p.added || p.token == "" {
		if err == nil {
			t.Error("no error refreshing invalid player / token")
		}
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	token := got.GetToken()
	if len(token) == 0 {
		t.Error("refresh token empty")
		return
	}
	p.token = token

}

func (ts *testServer) randomAction(t *testing.T, p *testPlayer) {
	ts.actions[int64(math.Floor(ts.r.Float64()*float64(len(ts.actions))))](t, p)
}

func comparePlayers(t *testing.T, tp *testPlayer, mp *messages.Player) {
	if tp == nil || mp == nil {
		t.Error("both players are non-nil")
		return
	}
	if mp.GetUsername() != tp.username {
		t.Errorf("players have mismatched username, got: %s, want: %s", mp.GetUsername(), tp.username)
		return
	}
}
