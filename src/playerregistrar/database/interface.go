package database

import (
	"crypto/rsa"
	"flag"
	"fmt"
	"strings"

	"github.com/sambdavidson/community-chess/src/proto/messages"
)

var (
	dbKind = flag.String("database_kind", "memory", `The kind of database to use. Supported strings are:
- 'memory': an ephemeral in-memory database for use in testing
- 'postgres': a postgres database, requires a running db with the correct tables/schema. Run with --help to see all postgres flags.`)
)

// Database is the interface by which one can interact with a database.
type Database interface {
	RegisterPlayer(string) (*messages.Player, error)
	GetPlayerByID(string) (*messages.Player, error)
	GetPlayerByUsername(string, int32) (*messages.Player, error)
	GetAllValidKeys() ([]*messages.TimedPrivateKey, error)
	AddKey(key *rsa.PrivateKey, validSeconds int64) error
	Close()
}

var (
	defaultInstance Database
)

// DefaultInstance returns the default singleton database based on the --database_kind flag.
func DefaultInstance() (Database, error) {
	if defaultInstance == nil {
		db, err := Instance(*dbKind)
		if err != nil {
			return nil, err
		}
		defaultInstance = db
	}
	return defaultInstance, nil
}

// Instance returns a singleton playerregistrar database object based on the kind parameter.
func Instance(kind string) (Database, error) {
	switch strings.ToLower(kind) {
	case "memory":
		return memoryInstance(), nil
	case "postgres":
		return postgresInstance()
	}
	return nil, fmt.Errorf("unsupported database type: %s", *dbKind)
}
