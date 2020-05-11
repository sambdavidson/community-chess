package database

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sambdavidson/community-chess/src/proto/messages"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type memoryDB struct {
	mux sync.RWMutex

	players             map[string]*messages.Player
	usernameCounts      map[string]int32
	usernameNumbersToID map[string]map[int32]string

	keyMux sync.RWMutex
	keys   []*messages.TimedPrivateKey
}

func memoryInstance() Database {
	db := &memoryDB{
		players:             map[string]*messages.Player{},
		usernameCounts:      map[string]int32{},
		usernameNumbersToID: map[string]map[int32]string{},
	}
	return db
}

// RegisterPlayer adds a Player to the database, returns an error if something went wrong.
func (db *memoryDB) RegisterPlayer(name string) (*messages.Player, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	count := db.usernameCounts[name] + 1
	if count > 9999 {
		return nil, status.Error(codes.ResourceExhausted, "Username all used up.")
	}

	pid := uuid.New().String()
	p := &messages.Player{
		Id:           pid,
		CreationTime: time.Now().UnixNano(),
		NumberSuffix: count,
		Username:     name,
	}

	if countToID, ok := db.usernameNumbersToID[name]; ok {
		countToID[count] = pid
	} else {
		db.usernameNumbersToID[name] = map[int32]string{count: pid}
	}

	db.players[pid] = p
	return p, nil
}

// Get a Player by their UUID ID, returns that player and true if found, false and an error if not or an error occured.
func (db *memoryDB) GetPlayerByID(id string) (*messages.Player, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	player, ok := db.players[id]
	if !ok {
		return nil, status.Error(codes.NotFound, "Unknown player")
	}
	return player, nil
}

func (db *memoryDB) GetPlayerByUsername(username string, suffix int32) (*messages.Player, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	id, ok := db.usernameNumbersToID[username][suffix]
	if !ok {
		return nil, nil
	}
	return db.GetPlayerByID(id)
}

// GetAllValidKeys returns the in-memory set of TimedPrivateKeys.
func (db *memoryDB) GetAllValidKeys() ([]*messages.TimedPrivateKey, error) {
	db.keyMux.RLock()
	defer db.keyMux.RUnlock()
	return db.keys, nil
}

// AddKey adds a new TimedPrivateKey to the in-memory set.
func (db *memoryDB) AddKey(key *rsa.PrivateKey, validSeconds int64) error {
	db.keyMux.Lock()
	defer db.keyMux.Unlock()
	db.keys = append(db.keys, &messages.TimedPrivateKey{
		KeyId:        int64(len(db.keys)),
		Iss:          int64(time.Now().Second()),
		ValidSeconds: validSeconds,
		PemPrivateKey: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		}),
	})
	return nil
}

// Close the connection to the database.
func (db *memoryDB) Close() {}
