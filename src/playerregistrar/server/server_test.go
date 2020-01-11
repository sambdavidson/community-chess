package server

import (
	"math"
	"math/rand"
	"testing"

	"github.com/google/uuid"
)

func TestServerE2E(t *testing.T) {
	playerCount := 10000
	players := make([]struct {
		id        string
		token     string
		added     bool
		got       bool
		refreshed bool
	}, playerCount)
	r := rand.New(rand.NewSource(1))
	s, err := New(&Opts{})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < (playerCount * 10); i++ {
		p := players[int64(math.Floor(r.Float64()*float64(playerCount)))]
		switch {
		case !p.added:
			pid, err := uuid.NewRandomFromReader(r)
			if err != nil {
				t.Fatal(err)
			}
			p.id = pid
			p.added = true

		}
	}
}
