package playertoken

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/sambdavidson/community-chess/src/playerregistrar/database"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	keyer           = flag.Bool("keyer", false, "Whether this playerregistrar/playertoken instance should append new keys to the DB.")
	addKeyOnEmptyDB = flag.Bool("add_key_on_empty_db", false, "Whether to add a first key to the DB if it is empty. Requires the --keyer flag to be set.")
)

const (
	defaultTTL = time.Minute * 30
)

// Issuer issues player tokens in exchange for players.
type Issuer struct {
	mux sync.RWMutex

	db       database.Database
	tokenTTL time.Duration
	keys     *keys
}

// IssuerOpts provides options for an Issuer object.
type IssuerOpts struct {
	DB       database.Database
	TokenTTL time.Duration
}

// NewTokenIssuer returns a new Issuer struct configured with the IssuerOpts.
func NewTokenIssuer(opts *IssuerOpts) (*Issuer, error) {
	if opts == nil {
		return nil, fmt.Errorf("token issuer options cannot be nil")
	}
	if opts.DB == nil {
		return nil, fmt.Errorf("token issuer database in options cannot be nil")
	}
	if opts.TokenTTL <= 0 {
		log.Printf("token issuer options TTL is %s, using default TTL of %s\n",
			opts.TokenTTL, defaultTTL)
		opts.TokenTTL = defaultTTL
	}
	i := &Issuer{
		db:       opts.DB,
		tokenTTL: opts.TokenTTL,
		keys:     &keys{},
	}
	return i, i.UpdateKeys()
}

// PublicKeys gives all the public keys
func (i *Issuer) PublicKeys() []*messages.TimedPublicKey {
	i.mux.RLock()
	defer i.mux.RUnlock()
	return i.keys.PublicKeys()
}

// TokenForPlayer exchanged a player for a signed player token JWT string.
func (i *Issuer) TokenForPlayer(p *messages.Player) (string, error) {
	i.mux.RLock()
	defer i.mux.RUnlock()

	id, key := i.keys.SigningKey()
	if key == nil {
		return "", fmt.Errorf("missing signing key")
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, &jwt.StandardClaims{
		Issuer:    strconv.FormatInt(id, 10),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(i.tokenTTL).Unix(),
		Subject:   p.GetId(),
	})
	return token.SignedString(key)
}

// UpdateKeys fetches all valid keys from the database and updates the local set
// of keys to sign distribute public keys of.
func (i *Issuer) UpdateKeys() error {
	keys, err := i.db.GetAllValidKeys()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		if *addKeyOnEmptyDB {
			if err = i.newKey(false); err != nil {
				return err
			}
			keys, err := i.db.GetAllValidKeys()
			if err != nil {
				return err
			}
			if len(keys) > 0 {
				return i.keys.SetKeys(keys)
			}
		}
		return fmt.Errorf("empty list of player token keys from database")
	}
	return i.keys.SetKeys(keys)
}

// newKey adds a new RSA key to the DB, refreshes the keys, and starts using the latest key.
func (i *Issuer) newKey(update bool) error {
	if !*keyer {
		return fmt.Errorf("this playertoken.Issuer does not have the --keyer flag enabled")
	}
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	if err = i.db.AddKey(pk, int64(defaultTTL.Seconds())); err != nil {
		return err
	}
	return i.UpdateKeys()
}
