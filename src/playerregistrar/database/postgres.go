package database

import (
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"flag"
	"fmt"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	// Postgres support for database/sql
	_ "github.com/lib/pq"
)

// Flags
var (
	host = flag.String("postgres_host", "localhost", "Host of Postgres database (do not include port, use --postgres_port).")
	port = flag.String("postgres_port", "5433", "Host port of Postgres database.")
	// TODO specify other stuff like username, password, database name, ssl maybe.
)

type postgresDB struct {
	*sql.DB
}

func postgresInstance() (*postgresDB, error) {
	connectionStr := fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=community_chess sslmode=disable",
		*host, *port)
	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		return nil, err
	}
	return &postgresDB{db}, nil
}

func (db *postgresDB) RegisterPlayer(username string) (*messages.Player, error) {
	suffix, err := db.reserveNextUsernameSuffix(username)
	if err != nil {
		return nil, err
	}
	return scanRowIntoPlayer(db.QueryRow("INSERT INTO public.players VALUES (uuid_generate_v4(), $1, $2) RETURNING *;", username, suffix))
}
func (db *postgresDB) GetPlayerByID(id string) (*messages.Player, error) {
	rows, err := db.Query("SELECT * FROM public.players WHERE id=$1", id)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, nil
	}
	return scanRowIntoPlayer(rows)
}

func (db *postgresDB) GetPlayerByUsername(username string, suffix int32) (*messages.Player, error) {
	rows, err := db.Query("SELECT * FROM public.players WHERE username=$1 AND number_suffix=$2", username, suffix)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, nil
	}
	return scanRowIntoPlayer(rows)
}

// GetAllValidKeys returns all valid player signing keys from the DB.
func (db *postgresDB) GetAllValidKeys() ([]*messages.TimedPrivateKey, error) {
	rows, err := db.Query("SELECT key_id, iss_seconds, valid_seconds, key_pem FROM public.playertoken_keys WHERE expires_at_seconds > $1;",
		time.Now().Unix(),
	)
	if err != nil {
		return nil, err
	}
	var out []*messages.TimedPrivateKey
	for rows.Next() {
		tpk := &messages.TimedPrivateKey{}
		err := rows.Scan(
			&tpk.KeyId,
			&tpk.Iss,
			&tpk.ValidSeconds,
			&tpk.PemPrivateKey,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, tpk)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

// AddKey converts the passed RSA key to a TimedPrivateKey and adds it to the
// postgress DB.
func (db *postgresDB) AddKey(key *rsa.PrivateKey, validSeconds int64) error {
	_, err := db.Query(
		"INSERT INTO public.playertoken_keys (iss_seconds, valid_seconds, key_pem) VALUES ($1, $2, $3);",
		time.Now().Unix(),
		validSeconds,
		pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		}),
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *postgresDB) Close() {
	db.Close()
}

func (db *postgresDB) reserveNextUsernameSuffix(username string) (int32, error) {
	if err := db.ensureUsernameUsageRowExists(username); err != nil {
		return 0, err
	}
	row := db.QueryRow("UPDATE public.username_usage SET count = count + 1 WHERE username = $1 RETURNING count;", username)
	var suffix int32
	if err := row.Scan(&suffix); err != nil {
		return 0, err
	}
	return suffix, nil
}

func (db *postgresDB) ensureUsernameUsageRowExists(username string) error {
	_, err := db.Query("INSERT INTO public.username_usage VALUES ($1, 0) ON CONFLICT DO NOTHING;", username)
	fmt.Println("ENSURE ERROR", err)
	return err
}

// Accepts both *sql.Row and *sql.Rows.
type scanable interface {
	Scan(...interface{}) error
}

func scanRowIntoPlayer(row scanable) (*messages.Player, error) {
	player := &messages.Player{}
	if err := row.Scan(
		&player.Id,
		&player.Username,
		&player.NumberSuffix,
	); err != nil {
		return nil, err
	}
	return player, nil
}
