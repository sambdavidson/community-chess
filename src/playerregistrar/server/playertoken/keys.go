package playertoken

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/sambdavidson/community-chess/src/proto/messages"
)

type keys struct {
	mux         sync.RWMutex
	privateKeys []*timedPrivateKey
	publicKeys  []*messages.TimedPublicKey
}

// timedPrivateKey is a RSA private key that should not be used before or after
// a given time.
type timedPrivateKey struct {
	proto         *messages.TimedPrivateKey
	parsedPrivate *rsa.PrivateKey
}

// SetKeys sets the keys object with this set of TimedPrivateKeys. Returns an
// error if anything goes wrong parsing the keys.
func (k *keys) SetKeys(tks []*messages.TimedPrivateKey) error {
	timedPrivKeys := make([]*timedPrivateKey, len(tks))
	timedPubKeys := make([]*messages.TimedPublicKey, len(tks))
	for i, key := range tks {
		// Parse Private Key
		p, _ := pem.Decode(key.GetPemPrivateKey())

		pk, err := x509.ParsePKCS1PrivateKey(p.Bytes)
		if err != nil {
			return fmt.Errorf("unable to parse key_id: %d from database: %v", key.GetKeyId(), err)
		}
		pk.Precompute()
		// Encode Public Key
		timedPrivKeys[i] = &timedPrivateKey{
			proto:         key,
			parsedPrivate: pk,
		}
		timedPubKeys[i] = &messages.TimedPublicKey{
			KeyId:        key.GetKeyId(),
			Iss:          key.GetIss(),
			ValidSeconds: key.GetValidSeconds(),
			PemPublicKey: pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: x509.MarshalPKCS1PublicKey(&pk.PublicKey),
			}),
		}
	}
	k.mux.Lock()
	defer k.mux.Unlock()
	k.privateKeys = timedPrivKeys
	k.publicKeys = timedPubKeys
	return nil
}

// SigningKeys returns KeyID and rsa.Private key for signing player tokens.
// Returns 0, nil if there does not exist a signing key.
func (k *keys) SigningKey() (int64, *rsa.PrivateKey) {
	if len(k.privateKeys) == 0 {
		return 0, nil
	}
	key := k.privateKeys[len(k.privateKeys)-1]
	return key.proto.GetKeyId(), key.parsedPrivate
}

// PublicKeys returns the slice of known public keys which should be used for
// validating playertokens.
func (k *keys) PublicKeys() []*messages.TimedPublicKey {
	return k.publicKeys
}
