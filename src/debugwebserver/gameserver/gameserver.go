package gameserver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages"
	"github.com/sambdavidson/community-chess/src/proto/messages/games"

	"github.com/sambdavidson/community-chess/src/lib/auth/grpcplayertokens"

	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Handler handles gameserver operation HTTP stuff.
type Handler struct {
	TLS *tls.Config
}

var (
	conn              *grpc.ClientConn
	gsc               gs.GameServerClient
	errorNotConnected = fmt.Errorf("not not connected to Game Server service")
)

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(100000)
	rw.Header().Set("Content-Type", "application/json")

	switch req.URL.Path { // Handlers for following the /games/ prefix e.g. the /games/join URL
	case "connect":
		h.connect(rw, req)
	case "connectionstatus":
		h.connectionstatus(rw, req)
	case "game":
		h.game(rw, req)
	case "metadata":
		h.metadata(rw, req)
	case "state":
		h.state(rw, req)
	case "history":
		h.history(rw, req)
	case "join":
		h.join(rw, req)
	case "leave":
		h.leave(rw, req)
	case "postvote":
		h.postvote(rw, req)
	case "status":
		h.status(rw, req)
	default:
		rw.WriteHeader(http.StatusNotFound)
	}
}

func (h *Handler) connect(rw http.ResponseWriter, req *http.Request) {
	if conn != nil {
		conn.Close()
	}
	var err error
	conn, err = grpc.Dial(req.FormValue("gs-connect-address"),
		grpc.WithTransportCredentials(credentials.NewTLS(h.TLS)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			PermitWithoutStream: true,
			Time:                time.Hour,
			Timeout:             10 * time.Second,
		}))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(struct {
		Target string
		State  string
	}{conn.Target(), conn.GetState().String()})
	gsc = gs.NewGameServerClient(conn)
}

func (h *Handler) connectionstatus(rw http.ResponseWriter, req *http.Request) {
	target := ""
	state := "NOT CONNECTED"
	if conn != nil {
		target = conn.Target()
		state = conn.GetState().String()
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(struct {
		Target string
		State  string
	}{target, state})
}

func (h *Handler) game(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	ctx, err := ctxWithPToken(req.FormValue("player-token"))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
	}
	res, err := gsc.Game(ctx, &gs.GameRequest{
		Detailed: req.FormValue("gs-game-detailed") == "false",
	})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}

func (h *Handler) metadata(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	res, err := gsc.Metadata(context.Background(), &gs.MetadataRequest{})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}

func (h *Handler) state(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	res, err := gsc.State(context.Background(), &gs.StateRequest{
		Detailed: req.FormValue("gs-state-detailed") == "false",
	})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}
func (h *Handler) history(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	res, err := gsc.History(context.Background(), &gs.HistoryRequest{
		Detailed: req.FormValue("gs-history-detailed") == "false",
	})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}
func (h *Handler) join(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	ctx, err := ctxWithPToken(req.FormValue("player-token"))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
	}
	res, err := gsc.Join(ctx, &gs.JoinRequest{
		Fields: &messages.Game_NewPlayerFields{
			Game: &messages.Game_NewPlayerFields_ChessFields{
				ChessFields: &games.ChessNewPlayerFields{
					WhiteTeam: req.FormValue("gs-join-team") != "false",
				},
			},
		},
	})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}
func (h *Handler) leave(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	ctx, err := ctxWithPToken(req.FormValue("player-token"))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
	}
	res, err := gsc.Leave(ctx, &gs.LeaveRequest{})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}
func (h *Handler) postvote(rw http.ResponseWriter, req *http.Request) {

	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}
func (h *Handler) status(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}

func ctxWithPToken(t string) (context.Context, error) {
	if len(t) < 20 {
		return nil, fmt.Errorf("bad player id: %s", t)
	}
	ctx, err := grpcplayertokens.AppendPlayerAuthToOutgoingContext(context.Background(), t)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}
