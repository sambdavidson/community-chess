package playertokens

import (
	"context"
	"math"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

func TestAppendingPlayerTokensToOutgoingContext(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		ctx     context.Context
		token   string
		wantErr bool
	}{
		{
			"happy",
			metadata.NewOutgoingContext(context.TODO(), metadata.MD{}),
			"playerToken",
			false,
		},
		{
			"token already exists",
			metadata.NewOutgoingContext(context.TODO(), metadata.MD{
				playerTokenKey: []string{"existingPlayerToken"},
			}),
			"playerToken",
			true,
		},
		{
			"invalid non-grpc context",
			context.TODO(),
			"playerToken",
			true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			gotCtx, err := AppendPlayerAuthToOutgoingContext(tc.ctx, tc.token)
			if err != nil {
				if !tc.wantErr {
					t.Fatalf("got unwated error: %v", err)
				}
			} else {
				md, ok := metadata.FromOutgoingContext(gotCtx)
				if !ok {
					t.Fatalf("unable to get gotCtx metadata")
				}
				if got, ok := md[playerTokenKey]; ok {
					if len(got) != 1 {
						t.Fatalf("metadata for player token should contain exactly 1 value: %v", got)
					}
					if got[0] != tc.token {
						t.Fatalf("got: %s; want: %s", got[0], tc.token)
					}
				} else {
					t.Fatalf("gotCtx metadata missing value for key: %s", playerTokenKey)
				}
			}
		})
	}
}

func TestIngressUnaryInterceptor(t *testing.T) {
	for _, tc := range []struct {
		desc string
		// struct def
	}{
		// test cases
	} {
		// test runner
		t.Run(tc.desc, func(t *testing.T) {
			t.Skip()
			// TODO: Use the mock and build out tests for the interceptor.
			// You will need to do some crypto auth stuff using the JWT library.
		})
	}
}

func TestValidatedPlayerIDFromIncomingContext(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		ctx     context.Context
		want    string
		wantErr bool
	}{
		{
			"happy",
			metadata.NewIncomingContext(context.TODO(), metadata.MD{
				playerValidatedID: []string{"validatedPlayerIDFoo"},
			}),
			"validatedPlayerIDFoo",
			false,
		},
		{
			"missing metadata",
			context.TODO(),
			"",
			true,
		},
		{
			"missing player ID",
			metadata.NewIncomingContext(context.TODO(), metadata.MD{}),
			"",
			true,
		},
		{
			"invalid player ID",
			metadata.NewIncomingContext(context.TODO(), metadata.MD{
				playerValidatedID: []string{},
			}),
			"",
			true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got, gotErr := ValidatedPlayerIDFromIncomingContext(tc.ctx)
			if gotErr != nil {
				if !tc.wantErr {
					t.Fatalf("got unwanted error: %v", gotErr)
				}
			}
			if got != tc.want {
				t.Fatalf("got: %s; want %s", got, tc.want)
			}
		})
	}
}

func TestAutoRefresh(t *testing.T) {
	cli := &mockPlayerRegistrarClient{
		keys: []*registrar.TokenKeysResponse_TimeToKey{
			&registrar.TokenKeysResponse_TimeToKey{
				Key:       "abc",
				NotBefore: 0,
				NotAfter:  math.MaxInt64,
			},
		},
	}
	ing := NewPlayerAuthIngress(PlayerAuthIngressArgs{
		AutoRefreshCadence:     time.Millisecond * 250,
		PlayersRegistrarClient: cli,
	}).(*playerAuthIngress)
	time.Sleep(time.Second)
	cli.keys[0].Key = "123" // Update key and see if that is reflected
	time.Sleep(time.Second)
	ing.ticker.Stop()
	if cli.calls < 8 || cli.calls > 9 {
		t.Errorf("expected player registrar client to be called either 8 or 9 times, instead was called %d times", cli.calls)
	}

	if len(ing.keys) != 1 && ing.keys[0].Key != "123" {
		t.Errorf("invalid ending keys, got: %v; want: %v", ing.keys, cli.keys)
	}
}

/* MOCKS */
var (
	unimplementedErr = status.Error(codes.Unimplemented, "not implemented")
)

type mockPlayerRegistrarClient struct {
	calls int
	keys  []*registrar.TokenKeysResponse_TimeToKey
}

func (c *mockPlayerRegistrarClient) RegisterPlayer(ctx context.Context, in *registrar.RegisterPlayerRequest, opts ...grpc.CallOption) (*registrar.RegisterPlayerResponse, error) {
	return nil, unimplementedErr
}
func (c *mockPlayerRegistrarClient) GetPlayer(ctx context.Context, in *registrar.GetPlayerRequest, opts ...grpc.CallOption) (*registrar.GetPlayerReponse, error) {
	return nil, unimplementedErr
}
func (c *mockPlayerRegistrarClient) Login(ctx context.Context, in *registrar.LoginRequest, opts ...grpc.CallOption) (*registrar.LoginResponse, error) {
	return nil, unimplementedErr
}
func (c *mockPlayerRegistrarClient) RefreshToken(ctx context.Context, in *registrar.RefreshTokenRequest, opts ...grpc.CallOption) (*registrar.RefreshTokenResponse, error) {
	return nil, unimplementedErr
}
func (c *mockPlayerRegistrarClient) TokenKeys(ctx context.Context, in *registrar.TokenKeysRequest, opts ...grpc.CallOption) (*registrar.TokenKeysResponse, error) {
	c.calls++
	return &registrar.TokenKeysResponse{
		History: c.keys,
	}, nil
}
