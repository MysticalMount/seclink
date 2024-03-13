package db

import (
	"path/filepath"
	"seclink/log"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/spf13/viper"
)

type ISeclinkDb interface {
	Start(lock bool, ro bool) error
	Get([]byte) ([]byte, error)
	Set([]byte, []byte, time.Duration) error
	Close() error
}

type SSeclinkDb struct {
	db *badger.DB
}

func (d *SSeclinkDb) Start(lock bool, ro bool) error {
	l := log.Get()

	dbPath := filepath.Join(viper.GetString("server.datapath"), "db")
	l.Info().
		Str("DbPath", dbPath).
		Msg("Attempting to open BadgerDB")

	db, err := badger.Open(badger.DefaultOptions(dbPath).WithBypassLockGuard(lock).WithReadOnly(ro))
	if err != nil {
		l.Error().
			Err(err).
			Msg("An error was encountered opening the BadgerDB")
		return err
	}

	// Success, assign the db to the struct
	d.db = db
	l.Info().Msg("Successfully started the BadgerDB")
	return nil
}

// Closes the DB
func (d *SSeclinkDb) Close() error {
	return d.db.Close()
}

// Retrieves a key from the db
func (d *SSeclinkDb) Get(key []byte) ([]byte, error) {
	var value []byte
	err := d.db.View(func(txn *badger.Txn) error {

		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			// This func with val would only be called if item.Value encounters no error.

			value = val

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})

	// Final handle
	if err != nil {
		return nil, err
	} else {
		return value, err
	}
}

// Sets a key in the db
func (d *SSeclinkDb) Set(key []byte, val []byte, ttl time.Duration) error {
	l := log.Get()

	l.Trace().Bytes("id", key).Bytes("val", val).Dur("ttl", ttl).Msg("Set trace")
	err := d.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, val).WithTTL(ttl)
		err := txn.SetEntry(e)
		return err
	})
	return err
}

// New Seclink DB
func NewSeclinkDb() ISeclinkDb {
	return &SSeclinkDb{}
}
