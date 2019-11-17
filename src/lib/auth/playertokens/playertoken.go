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

// PlayerAuthIngress is an object that validates incoming Player token ingress.
type PlayerAuthIngress interface {
	TokenValidationUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
	ValidatedPlayerIDFromContext(ctx context.Context) (string, error)
}

// PlayerAuthIngressArgs are the arguments for a new PlayerAuthIngress
type PlayerAuthIngressArgs struct {
	playersRegistrarClient registrar.PlayersRegistrarClient
}

// NewPlayerAuthIngress builds a new PlayerAuthIngress
func NewPlayerAuthIngress(args PlayerAuthIngressArgs) PlayerAuthIngress {
	p := &playerAuthIngress{
		playersRegistrarClient: args.playersRegistrarClient,
		keys:                   []*registrar.TokenKeysResponse_TimeToKey{},
	}
	return p
}

type playerAuthIngress struct {
	playersRegistrarClient registrar.PlayersRegistrarClient

	mux  sync.Mutex
	keys []*registrar.TokenKeysResponse_TimeToKey
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

func (p *playerAuthIngress) ValidatedPlayerIDFromContext(ctx context.Context) (string, error) {
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

func (p *playerAuthIngress) keyForToken(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
	}

	c, ok := t.Claims.(jwt.StandardClaims)
	if !ok {
		return nil, fmt.Errorf("player token claims are incorrect type")
	}

	// Keys are sorted to newest to latest, so the first key we are after the NotBefore should be correct.
	for _, k := range p.keys {
		if c.IssuedAt > k.GetNotBefore() {
			// TODO: locate the best key for this player token.
			// Maybe RPC to get the latest keys if its missing.
		}
	}

	// TODO: use the PR client to get internal keys for this.
	return "foobar", nil
}

func (p *playerAuthIngress) refreshSecrets() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := p.playersRegistrarClient.TokenKeys(ctx, &registrar.TokenKeysRequest{})
	if err != nil {
		log.Printf("error: unable to refresh player token ingress secrets: %v", err)
		return
	}
	if len(res.GetHistory()) == 0 {
		log.Printf("error: bad player token ingress secrets, history is empty")
	}
	p.mux.Lock()
	defer p.mux.Unlock()
	p.keys = res.GetHistory()

}
