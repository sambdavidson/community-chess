package grpcplayertokens

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"

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
	pk1 := privateRSA(t)
	pk2 := privateRSA(t)

	wantPID := "testPlayerId"
	now := time.Unix(100, 0)
	jwt.TimeFunc = func() time.Time { return now } // Stub out time.Now such that we can test invalid certs.

	type ttpk = registrar.TokenPublicKeysResponse_TimeToPublicKey

	jwtSignedString := func(pk *rsa.PrivateKey, claims *jwt.StandardClaims) string {
		j := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
		ss, err := j.SignedString(pk)
		if err != nil {
			t.Fatal(err)
		}
		return ss
	}

	for _, tc := range []struct {
		desc             string
		initHistory      []*registrar.TokenPublicKeysResponse_TimeToPublicKey
		registrarHistory []*registrar.TokenPublicKeysResponse_TimeToPublicKey
		jwt              string
		wantErr          bool
	}{
		{
			desc: "happy init history has pk",
			initHistory: []*ttpk{
				historyKey(t, pk1, 0, 50),
			},
			registrarHistory: []*ttpk{
				historyKey(t, pk1, 0, 50),
			},
			jwt: jwtSignedString(pk1, &jwt.StandardClaims{
				IssuedAt:  10,
				NotBefore: 10,
				ExpiresAt: 150,
				Subject:   wantPID,
			}),
			wantErr: false,
		},
		{
			desc:        "happy query for missing history",
			initHistory: []*ttpk{},
			registrarHistory: []*ttpk{
				historyKey(t, pk1, 0, 50),
			},
			jwt: jwtSignedString(pk1, &jwt.StandardClaims{
				IssuedAt:  10,
				NotBefore: 10,
				ExpiresAt: 150,
				Subject:   wantPID,
			}),
			wantErr: false,
		},
		{
			desc: "happy new key query for new history",
			initHistory: []*ttpk{
				historyKey(t, pk1, 0, 50),
			},
			registrarHistory: []*ttpk{
				historyKey(t, pk2, 50, 90),
				historyKey(t, pk1, 0, 50),
			},
			jwt: jwtSignedString(pk2, &jwt.StandardClaims{
				IssuedAt:  60,
				NotBefore: 60,
				ExpiresAt: 200,
				Subject:   wantPID,
			}),
			wantErr: false,
		},
		{
			desc: "sad missing jwt",
			initHistory: []*ttpk{
				historyKey(t, pk1, 0, 50),
			},
			registrarHistory: []*ttpk{
				historyKey(t, pk1, 0, 50),
			},
			jwt:     "",
			wantErr: true,
		},
		{
			desc: "sad bad key",
			initHistory: []*ttpk{
				historyKey(t, pk1, 0, 50),
			},
			registrarHistory: []*ttpk{
				historyKey(t, pk1, 0, 50),
			},
			jwt: jwtSignedString(pk2, &jwt.StandardClaims{ // pk2
				IssuedAt:  10,
				NotBefore: 10,
				ExpiresAt: 150,
				Subject:   wantPID,
			}),
			wantErr: true,
		},
		{
			desc: "sad key out of range",
			initHistory: []*ttpk{
				historyKey(t, pk1, 0, 50),
			},
			registrarHistory: []*ttpk{
				historyKey(t, pk1, 0, 50),
			},
			jwt: jwtSignedString(pk1, &jwt.StandardClaims{ // pk2
				IssuedAt:  60,
				NotBefore: 60,
				ExpiresAt: 150,
				Subject:   wantPID,
			}),
			wantErr: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			ing := &playerAuthIngress{
				playersRegistrarClient: &mockPlayerRegistrarClient{
					keys: tc.registrarHistory,
				},
				keys: tc.initHistory,
			}
			ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{
				playerTokenKey: []string{tc.jwt},
			})
			//	TokenValidationUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
			_, gotErr := ing.TokenValidationUnaryServerInterceptor(ctx,
				/* req */ nil,
				/* info */ nil,
				/* handler */ func(ctx context.Context, req interface{}) (interface{}, error) {
					// If the handler is called, the interceptor validated the request.
					gotPID, err := ValidatedPlayerIDFromIncomingContext(ctx)
					if err != nil {
						t.Errorf("handler error getting player ID from context: %v", err)
					}
					if gotPID != wantPID {
						t.Errorf("validated player ID incorrect got: %s; want: %s", gotPID, wantPID)
					}
					return nil, nil
				})
			if tc.wantErr {
				if gotErr == nil {
					t.Errorf("wanted non-nil error")
				}
			} else if gotErr != nil {
				t.Errorf("wanted nil error; got: %v", gotErr)
			}

		})
	}
}

func privateRSA(t *testing.T) *rsa.PrivateKey {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return pk
}

func historyKey(t *testing.T, pk *rsa.PrivateKey, notbefore, notafter int64) *registrar.TokenPublicKeysResponse_TimeToPublicKey {
	pubASN1, err := x509.MarshalPKIXPublicKey(&pk.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return &registrar.TokenPublicKeysResponse_TimeToPublicKey{
		PemPublicKey: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubASN1,
		}),
		NotBefore: notbefore,
		NotAfter:  notafter,
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
		keys: []*registrar.TokenPublicKeysResponse_TimeToPublicKey{
			&registrar.TokenPublicKeysResponse_TimeToPublicKey{
				PemPublicKey: []byte{0x1, 0x2, 0x3},
				NotBefore:    0,
				NotAfter:     math.MaxInt64,
			},
		},
	}
	ing := NewPlayerAuthIngress(PlayerAuthIngressArgs{
		AutoRefreshCadence:     time.Millisecond * 250,
		PlayersRegistrarClient: cli,
	}).(*playerAuthIngress)
	time.Sleep(time.Second)
	cli.keys[0].PemPublicKey = []byte{0x11, 0x22, 0x33} // Update key and see if that is reflected
	time.Sleep(time.Second)
	ing.ticker.Stop()
	if cli.calls < 8 || cli.calls > 9 {
		t.Errorf("expected player registrar client to be called either 8 or 9 times, instead was called %d times", cli.calls)
	}
	if len(ing.keys) != 1 && !reflect.DeepEqual(ing.keys[0].PemPublicKey, []byte{0x11, 0x22, 0x33}) {
		t.Errorf("invalid ending keys, got: %v; want: %v", ing.keys, cli.keys)
	}
}

/* MOCKS */
var (
	unimplementedErr = status.Error(codes.Unimplemented, "not implemented")
)

type mockPlayerRegistrarClient struct {
	calls int
	keys  []*registrar.TokenPublicKeysResponse_TimeToPublicKey
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
func (c *mockPlayerRegistrarClient) TokenPublicKeys(ctx context.Context, in *registrar.TokenPublicKeysRequest, opts ...grpc.CallOption) (*registrar.TokenPublicKeysResponse, error) {
	c.calls++
	return &registrar.TokenPublicKeysResponse{
		History: c.keys,
	}, nil
}
