package playertoken

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	jwt "github.com/dgrijalva/jwt-go"
	pb "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

const (
	defaultTTL = time.Minute * 30
)

// Issuer issues player tokens in exchange for players.
type Issuer struct {
	mux sync.RWMutex

	tokenTTL   time.Duration
	keys       []timedKey
	publicKeys []*pb.TokenPublicKeysResponse_TimeToPublicKey
}

// IssuerOpts provides options for an Issuer object.
type IssuerOpts struct {
	// TODO: DB source options
	TokenTTL time.Duration
}

// timedPrivateKey is a RSA private key that should not be used before or after a given time.
type timedKey struct {
	key       *rsa.PrivateKey
	notBefore time.Time
	notAfter  time.Time // Default (zero) time if still active.
}

// NewTokenIssuer returns a new Issuer struct configured with the IssuerOpts.
func NewTokenIssuer(opts *IssuerOpts) (*Issuer, error) {
	iss := &Issuer{
		tokenTTL: defaultTTL,
		keys:     []timedKey{},
	}
	if opts != nil {
		if opts.TokenTTL > 0 {
			iss.tokenTTL = opts.TokenTTL
		}
	}
	return iss, iss.NewKey()
}

// NewKey immediately stops using the existing keys and create a new RSA key
// pair to sign new player tokens.
func (i *Issuer) NewKey() error {
	i.mux.Lock()
	defer i.mux.Unlock()
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	now := time.Now()
	priv := timedKey{
		key:       pk,
		notBefore: now,

		// TODO: investigate if this is a good idea for now.
		notAfter: time.Unix(1<<63-1, 0), // Max time
	}
	pub, err := privateToPublic(priv)
	if err != nil {
		return err
	}
	if last := len(i.keys) - 1; last >= 0 {
		old := i.keys[last]
		old.notAfter = now
		oldPub := i.publicKeys[last]
		oldPub.NotAfter = now.Unix()
	}
	i.keys = append(i.keys, priv)
	i.publicKeys = append(i.publicKeys, pub)
	return nil
}

// PublicKeys gives all the public keys
func (i *Issuer) PublicKeys() []*pb.TokenPublicKeysResponse_TimeToPublicKey {
	i.mux.RLock()
	defer i.mux.RUnlock()
	return i.publicKeys
}

// TokenForPlayer exchanged a player for a signed player token JWT string.
func (i *Issuer) TokenForPlayer(p *messages.Player) (string, error) {
	i.mux.RLock()
	defer i.mux.RUnlock()
	if len(i.keys) == 0 {
		return "", fmt.Errorf("no keys")
	}
	now := time.Now()
	pk := i.keys[len(i.keys)-1].key

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, &jwt.StandardClaims{
		Issuer:    "playerRegistrarTODO",
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(i.tokenTTL).Unix(),
		Subject:   p.GetId(),
	})
	return token.SignedString(pk)
}

func privateToPublic(pk timedKey) (*pb.TokenPublicKeysResponse_TimeToPublicKey, error) {
	pubASN1, err := x509.MarshalPKIXPublicKey(&pk.key.PublicKey)
	if err != nil {
		return nil, err
	}

	return &pb.TokenPublicKeysResponse_TimeToPublicKey{
		NotAfter:  pk.notAfter.Unix(),
		NotBefore: pk.notBefore.Unix(),
		PemPublicKey: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubASN1,
		}),
	}, nil
}
