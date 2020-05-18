package grpcplayertokens

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/metadata"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sambdavidson/community-chess/src/proto/messages"
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
	GetUnaryServerInterceptor(failureMode ValidationFailureMode) grpc.UnaryServerInterceptor
}

// ValidationFailureMode is an enum for what to do if validation of player tokens fails
type ValidationFailureMode int

// Possible failures modes for validation failure
const (
	Reject = iota
	Ignore
)

type parsedKey struct {
	proto  *messages.TimedPublicKey
	parsed *rsa.PublicKey
}

// PlayerAuthIngressArgs are the arguments for a new PlayerAuthIngress
type PlayerAuthIngressArgs struct {
	PlayersRegistrarClient registrar.PlayersRegistrarClient
	AutoRefreshCadence     time.Duration
}

type playerAuthIngress struct {
	playersRegistrarClient registrar.PlayersRegistrarClient

	mux        sync.Mutex
	keys       map[int64]*parsedKey
	largestKID int64
	ticker     *time.Ticker
}

// NewPlayerAuthIngress builds a new PlayerAuthIngress. Refresh cadences < 5 seconds are ignored and set to 1 hour.
func NewPlayerAuthIngress(args PlayerAuthIngressArgs) PlayerAuthIngress {
	refreshCadence := args.AutoRefreshCadence
	if refreshCadence == 0 {
		refreshCadence = time.Hour
	}
	p := &playerAuthIngress{
		playersRegistrarClient: args.PlayersRegistrarClient,
		keys:                   map[int64]*parsedKey{},
		largestKID:             -1, // Must be < 0 starting, because kID can be 0
		ticker:                 time.NewTicker(refreshCadence),
	}
	go func() {
		p.refreshPublicKeys()
		for range p.ticker.C {
			p.refreshPublicKeys()
		}
	}()
	return p
}

func (p *playerAuthIngress) GetUnaryServerInterceptor(failureMode ValidationFailureMode) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		err := p.tokenToMDSubject(ctx, req)
		if err != nil {
			if failureMode == Reject {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}

func (p *playerAuthIngress) tokenToMDSubject(ctx context.Context, req interface{}) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errMissingMetadata
	}
	md[playerValidatedID] = nil

	vals, ok := md[playerTokenKey]
	if !ok || len(vals) == 0 {
		return errMissingPlayerToken
	}
	t, err := jwt.ParseWithClaims(vals[0], &jwt.StandardClaims{}, p.keyForToken)
	if err != nil {
		debugLogf("jwt parse error: %v", err)
		return errBadPlayerToken
	}
	c, ok := t.Claims.(*jwt.StandardClaims)
	if !ok {
		debugLogf("unable to cast claims")
		return errBadPlayerToken
	}
	if err = c.Valid(); err != nil {
		debugLogf("claims invalid")
		return errBadPlayerToken
	}
	// Token is valid, write the validated player ID to context.
	md[playerValidatedID] = []string{c.Subject}
	return nil
}

func (p *playerAuthIngress) keyForToken(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
	}

	c, ok := t.Claims.(*jwt.StandardClaims)
	if !ok {
		return nil, fmt.Errorf("player token claims are incorrect type")
	}
	kID, err := strconv.ParseInt(c.Issuer, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing playertoken issuer%v", err)
	}
	return p.keyForID(kID)
}

func (p *playerAuthIngress) keyForID(id int64) (*rsa.PublicKey, error) {
	if id > p.largestKID {
		p.refreshPublicKeys()
	}
	key, ok := p.keys[id]
	if !ok {
		return nil, fmt.Errorf("no key for key_id: %v", id)
	}
	return key.parsed, nil
}

func (p *playerAuthIngress) refreshPublicKeys() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := p.playersRegistrarClient.TokenPublicKeys(ctx, &registrar.TokenPublicKeysRequest{})
	if err != nil {
		log.Printf("error: unable to refresh player token public keys: %v", err)
		return
	}
	history := res.GetHistory()

	keys := map[int64]*parsedKey{}
	var largestKID int64
	for _, hk := range history {
		block, _ := pem.Decode(hk.GetPemPublicKey())
		if block == nil {
			log.Printf("bad pem key data, no pem block")
			continue
		}
		pk, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			log.Printf("error: unable to marshal public key, key_id: %v", hk.GetKeyId())
			continue
		}
		keys[hk.GetKeyId()] = &parsedKey{
			proto:  hk,
			parsed: pk,
		}
		if largestKID < hk.GetKeyId() {
			largestKID = hk.GetKeyId()
		}
	}
	if len(keys) == 0 {
		log.Printf("error: bad player token public keys set is empty")
		return
	}
	p.mux.Lock()
	defer p.mux.Unlock()
	p.keys = keys
	p.largestKID = largestKID
}

func debugLogf(format string, v ...interface{}) {
	if f := flag.Lookup("debug"); f != nil && f.Value.(flag.Getter).Get().(bool) {
		log.Printf(format, v...)
	}
}
