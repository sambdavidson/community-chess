package playertokens

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/metadata"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

const (
	playerTokenKey    = "x-player-token"
	playerValidatedID = "x-player-validated-id"
)

// gRPC errors
var (
	errPlayerAuthAlreadyPresent = status.Errorf(codes.InvalidArgument, "player auth token already present")
	errMissingMetadata          = status.Errorf(codes.InvalidArgument, "missing metadata")
	errMissingPlayerToken       = status.Errorf(codes.PermissionDenied, "missing player auth token in metadata")
	errBadPlayerToken           = status.Errorf(codes.PermissionDenied, "player auth token is incorrectly formatted")
	errMissingValidatedPlayerID = status.Errorf(codes.Internal, "missing validated player ID")
	errBadPlayerID              = status.Errorf(codes.Internal, "player id is not a UUIDv4")
)

// AppendPlayerAuthToOutgoingContext appends the player auth token to the outgoing metadata in ctx.
// Returns an updated if successful; if the player ID is invalid or already exists an error is returned.
// The returned error is a valid GRPC error.
func AppendPlayerAuthToOutgoingContext(ctx context.Context, token string) (context.Context, error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		_, ok := md[playerTokenKey]
		if ok {
			return nil, errPlayerAuthAlreadyPresent
		}
	}
	return metadata.AppendToOutgoingContext(ctx, playerTokenKey, token), nil
}

// ValidatedPlayerIDFromIncomingContext returns a validated player ID from the passed incoming context. Returns an error if the
// passed context does not contain incoming metadata or if the validated player ID is misconfigured or missing. The validated
// player ID is populated by using a PlayerAuthIngress's TokenValidationUnaryServerInterceptor function within the GRPC service.
func ValidatedPlayerIDFromIncomingContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errMissingMetadata
	}
	vals, ok := md[playerValidatedID]
	if !ok || len(vals) == 0 {
		return "", errMissingValidatedPlayerID
	}
	return vals[0], nil
}

// PlayerAuthIngress is an object that validates incoming Player token ingress.
type PlayerAuthIngress interface {
	TokenValidationUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
}

// PlayerAuthIngressArgs are the arguments for a new PlayerAuthIngress
type PlayerAuthIngressArgs struct {
	PlayersRegistrarClient registrar.PlayersRegistrarClient
	AutoRefreshCadence     time.Duration
}

// NewPlayerAuthIngress builds a new PlayerAuthIngress
func NewPlayerAuthIngress(args PlayerAuthIngressArgs) PlayerAuthIngress {
	p := &playerAuthIngress{
		playersRegistrarClient: args.PlayersRegistrarClient,
		keys:                   []*registrar.TokenKeysResponse_TimeToKey{},
		ticker:                 time.NewTicker(args.AutoRefreshCadence),
	}
	go func() {
		p.refreshPublicKeys()
		for range p.ticker.C {
			p.refreshPublicKeys()
		}
	}()
	return p
}

type playerAuthIngress struct {
	playersRegistrarClient registrar.PlayersRegistrarClient

	mux    sync.Mutex
	keys   []*registrar.TokenKeysResponse_TimeToKey
	ticker *time.Ticker
}

func (p *playerAuthIngress) TokenValidationUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}
	vals, ok := md[playerTokenKey]
	if !ok || len(vals) == 0 {
		return nil, errMissingPlayerToken
	}
	t, err := jwt.Parse(vals[0], p.keyForToken)
	if err != nil {
		return nil, errBadPlayerToken
	}
	c, ok := t.Claims.(jwt.StandardClaims)
	if !ok {
		return nil, errBadPlayerToken
	}
	if err = c.Valid(); err != nil {
		return nil, errBadPlayerToken
	}

	// Token is valid, write the validated player ID to context.
	md[playerValidatedID] = []string{c.Subject}
	return handler(ctx, req)
}

func (p *playerAuthIngress) keyForToken(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
	}

	c, ok := t.Claims.(jwt.StandardClaims)
	if !ok {
		return nil, fmt.Errorf("player token claims are incorrect type")
	}

	return p.keyForTime(c.IssuedAt)
}

func (p *playerAuthIngress) keyForTime(iss int64) (string, error) {
	// Keys are sorted to newest to oldest, so the first key we are after the NotBefore should be correct.
	retry := true
	for retry {
		retry = false
		for _, k := range p.keys {
			if iss > k.GetNotBefore() {
				if iss > k.GetNotAfter() {
					p.refreshPublicKeys()
					retry = true
					break
				} else {
					return k.GetKey(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("missing key for time: %d", iss)
}

func (p *playerAuthIngress) refreshPublicKeys() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := p.playersRegistrarClient.TokenKeys(ctx, &registrar.TokenKeysRequest{})
	if err != nil {
		log.Printf("error: unable to refresh player token public keys: %v", err)
		return
	}
	if len(res.GetHistory()) == 0 {
		log.Printf("error: bad player token public keys, history is empty")
	}
	p.mux.Lock()
	defer p.mux.Unlock()
	p.keys = res.GetHistory()

}
