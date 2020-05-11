package playertoken

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/rand"
	"testing"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages"
)

var (
	currentTime time.Time
	count       int64
	rnd         = rand.New(rand.NewSource(12345))
)

func TestKeys(t *testing.T) {
	dbKeys := make([]*messages.TimedPrivateKey, 10)
	privKeys := make([]*rsa.PrivateKey, len(dbKeys))
	for i := 0; i < len(dbKeys); i++ {
		privKeys[i], dbKeys[i] = genTimedPrivateKey(t)
	}

	keys := keys{}

	t.Run("SetKeys", func(t *testing.T) {
		if err := keys.SetKeys(dbKeys); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("SigningKey", func(t *testing.T) {
		wantTPK := dbKeys[len(dbKeys)-1]
		wantPK := privKeys[len(dbKeys)-1]
		skID, sk := keys.SigningKey()
		if sk == nil {
			t.Fatal("got nil signing key")
		}
		if skID != wantTPK.GetKeyId() {
			t.Fatalf("unexpected signing key ID, got: %v; want: %v", skID, dbKeys[len(dbKeys)-1].GetKeyId())
		}
		if !keyEq(sk, wantPK) {
			t.Fatal("unexpected key, does not equal expected key values")
		}
	})

	t.Run("PublicKeys", func(t *testing.T) {
		pubs := keys.PublicKeys()
		if len(pubs) != len(dbKeys) {
			t.Fatalf("unexpected public keys slice length, got: %d; want: %d", len(pubs), len(dbKeys))
		}
	})

}

func TestEmptyKeys(t *testing.T) {
	keys := keys{}
	t.Run("SigningKey", func(t *testing.T) {
		if _, k := keys.SigningKey(); k != nil {
			t.Errorf("got: %v; want %v", k, nil)
		}
	})
	t.Run("PublicKeys", func(t *testing.T) {
		if pubs := keys.PublicKeys(); len(pubs) != 0 {
			t.Errorf("got: %v; want empty list", pubs)
		}
	})
}

func keyEq(k1 *rsa.PrivateKey, k2 *rsa.PrivateKey) bool {
	return k1.E == k2.E && k1.D.Cmp(k2.D) == 0 && k1.N.Cmp(k2.N) == 0
}

func getNowAndAdvanceTime() time.Time {
	t := currentTime
	currentTime = currentTime.Add(time.Hour * 10)
	return t
}

func genTimedPrivateKey(t *testing.T) (*rsa.PrivateKey, *messages.TimedPrivateKey) {
	pk, err := rsa.GenerateKey(rnd, 2048)
	if err != nil {
		t.Fatal(err)
	}
	id := count
	count++
	return pk, &messages.TimedPrivateKey{
		KeyId:        id,
		Iss:          getNowAndAdvanceTime().Unix(),
		ValidSeconds: int64(defaultTTL.Seconds()),
		PemPrivateKey: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(pk),
		}),
	}
}
