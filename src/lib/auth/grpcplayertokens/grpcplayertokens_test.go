package grpcplayertokens

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"reflect"
	"testing"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	jwt "github.com/dgrijalva/jwt-go"

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

	jwtSignedString := func(pk *rsa.PrivateKey, claims *jwt.StandardClaims) string {
		j := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
		ss, err := j.SignedString(pk)
		if err != nil {
			t.Fatal(err)
		}
		return ss
	}

	for _, tc := range []struct {
		desc          string
		initKeys      map[int64]*parsedKey
		registrarKeys []*messages.TimedPublicKey
		jwt           string
		wantErr       bool
	}{
		{
			desc: "happy init history has pk",
			initKeys: map[int64]*parsedKey{
				1: &parsedKey{
					proto:  keygen(t, pk1, 1, 0, 50),
					parsed: &pk1.PublicKey,
				},
			},
			registrarKeys: []*messages.TimedPublicKey{
				keygen(t, pk1, 1, 0, 500),
			},
			jwt: jwtSignedString(pk1, &jwt.StandardClaims{
				Issuer:    "1",
				IssuedAt:  10,
				NotBefore: 10,
				ExpiresAt: 150,
				Subject:   wantPID,
			}),
			wantErr: false,
		},
		{
			desc:     "happy query for missing history",
			initKeys: map[int64]*parsedKey{},
			registrarKeys: []*messages.TimedPublicKey{
				keygen(t, pk1, 1, 0, 50),
			},
			jwt: jwtSignedString(pk1, &jwt.StandardClaims{
				Issuer:    "1",
				IssuedAt:  10,
				NotBefore: 10,
				ExpiresAt: 150,
				Subject:   wantPID,
			}),
			wantErr: false,
		},
		{
			desc: "happy new key query for new history",
			initKeys: map[int64]*parsedKey{
				1: &parsedKey{
					proto:  keygen(t, pk1, 1, 0, 50),
					parsed: &pk1.PublicKey,
				},
			},
			registrarKeys: []*messages.TimedPublicKey{
				keygen(t, pk2, 2, 50, 90),
				keygen(t, pk1, 1, 0, 50),
			},
			jwt: jwtSignedString(pk2, &jwt.StandardClaims{
				Issuer:    "2",
				IssuedAt:  60,
				NotBefore: 60,
				ExpiresAt: 200,
				Subject:   wantPID,
			}),
			wantErr: false,
		},
		{
			desc: "sad missing jwt",
			initKeys: map[int64]*parsedKey{
				1: &parsedKey{
					proto:  keygen(t, pk1, 1, 0, 50),
					parsed: &pk1.PublicKey,
				},
			},
			registrarKeys: []*messages.TimedPublicKey{
				keygen(t, pk1, 1, 0, 50),
			},
			jwt:     "",
			wantErr: true,
		},
		{
			desc: "sad bad key",
			initKeys: map[int64]*parsedKey{
				1: &parsedKey{
					proto:  keygen(t, pk1, 1, 0, 50),
					parsed: &pk1.PublicKey,
				},
			},
			registrarKeys: []*messages.TimedPublicKey{
				keygen(t, pk1, 1, 0, 50),
			},
			jwt: jwtSignedString(pk2, &jwt.StandardClaims{
				Issuer:    "2",
				IssuedAt:  10,
				NotBefore: 10,
				ExpiresAt: 150,
				Subject:   wantPID,
			}),
			wantErr: true,
		},
		{
			desc: "sad key out of range",
			initKeys: map[int64]*parsedKey{
				1: &parsedKey{
					proto:  keygen(t, pk1, 1, 0, 50),
					parsed: &pk1.PublicKey,
				},
			},
			registrarKeys: []*messages.TimedPublicKey{
				keygen(t, pk1, 1, 0, 50),
			},
			jwt: jwtSignedString(pk1, &jwt.StandardClaims{
				Issuer:    "2",
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
					keys: tc.registrarKeys,
				},
				keys: tc.initKeys,
			}
			interceptor := ing.GetUnaryServerInterceptor(Reject)
			ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{
				playerTokenKey: []string{tc.jwt},
			})
			//	TokenValidationUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
			_, gotErr := interceptor(ctx,
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

func keygen(t *testing.T, pk *rsa.PrivateKey, id, iss, ttl int64) *messages.TimedPublicKey {
	pubASN1 := x509.MarshalPKCS1PublicKey(&pk.PublicKey)
	return &messages.TimedPublicKey{
		KeyId:        id,
		Iss:          iss,
		ValidSeconds: ttl,
		PemPublicKey: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubASN1,
		}),
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
	k1 := keygen(t, privateRSA(t), 0, 1000, 1000)
	k2 := keygen(t, privateRSA(t), 1, 2000, 1000)
	cli := &mockPlayerRegistrarClient{
		keys: []*messages.TimedPublicKey{k1},
	}
	ing := NewPlayerAuthIngress(PlayerAuthIngressArgs{
		AutoRefreshCadence:     time.Millisecond * 250,
		PlayersRegistrarClient: cli,
	}).(*playerAuthIngress)
	time.Sleep(time.Second)
	cli.keys = append(cli.keys, k2) // Update keys and see if that is reflected
	time.Sleep(time.Second)
	ing.ticker.Stop()
	if cli.calls < 8 || cli.calls > 9 {
		t.Errorf("expected player registrar client to be called either 8 or 9 times, instead was called %d times", cli.calls)
	}
	if len(ing.keys) != 2 ||
		!reflect.DeepEqual(ing.keys[0].proto, k1) ||
		!reflect.DeepEqual(ing.keys[1].proto, k2) {
		t.Errorf("invalid ending keys %v", ing.keys)
	}
}

/* MOCKS */
var (
	unimplementedErr = status.Error(codes.Unimplemented, "not implemented")
)

type mockPlayerRegistrarClient struct {
	calls int
	keys  []*messages.TimedPublicKey
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
