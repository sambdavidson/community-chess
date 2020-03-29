package database

import (
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

// Close the connection to the database.
func (db *memoryDB) Close() {}
