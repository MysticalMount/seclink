package db

import (
	"database/sql"
	"embed"
	"path/filepath"
	"seclink/log"

	_ "github.com/glebarez/go-sqlite"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/viper"
)

const databaseName = "seclink"
const databaseFilename = "sqlitedb"

//go:embed migrations
var migrations embed.FS

type ISeclinkDb interface {
	Start(lock bool, ro bool) error
	Migrate() error
	// GetAllLinks() ([]SSharedLink, error)
	Close() error
}

type SSeclinkDb struct {
	db *sql.DB
}

func (d *SSeclinkDb) Start(lock bool, ro bool) error {

	l := log.Get()

	dbPath := filepath.Join(viper.GetString("server.datapath"), databaseFilename)
	l.Info().
		Str("dbPath", dbPath).
		Msg("Attempting to open Sqlite database")

	// connect
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		l.Error().
			Err(err).
			Msg("An error was encountered opening the Sqlite database")
		return err
	}

	// Successfully connected
	l.Info().
		Msg("Sqlite connection established")

	d.db = db

	return nil
}

func (d *SSeclinkDb) Migrate() error {
	// NOTE: golang-migrate and sqlite3
	//
	// golang-migrate doesnt natively work with sqlite3 - as seclink has minimal data requirements in terms of performance, to avoid
	// these complications, seclink only uses sqlite (2)
	// In the future the work to use sqlite3 could be considered

	// Create golang-migrate database driver
	driver, err := sqlite.WithInstance(d.db, &sqlite.Config{DatabaseName: databaseName})

	// Initialise source from embedded migrations
	source, err := iofs.New(migrations, "migrations")

	// Create new golang-migrate instance
	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)

	// Migrate all
	m.Up()

	return err
}

// Closes the DB
func (d *SSeclinkDb) Close() error {
	return d.db.Close()
}

// Gets all keys in the db
// func (d *SSeclinkDb) GetAllLinks() ([]SSharedLink, error) {
// 	//l := log.Get()

// 	results := make([]SSharedLink, 0)

// 	err := d.db.View(func(txn *badger.Txn) error {
// 		opts := badger.DefaultIteratorOptions
// 		opts.PrefetchSize = 10
// 		it := txn.NewIterator(opts)
// 		defer it.Close()
// 		for it.Rewind(); it.Valid(); it.Next() {
// 			newResult := SSharedLink{}
// 			item := it.Item()
// 			k := item.Key()

// 			// Get timestamp
// 			expiresAt := time.Unix(int64(item.ExpiresAt()), 0)
// 			newResult.ExpiresAt = expiresAt
// 			newResult.Ttl = time.Duration(time.Since(expiresAt))
// 			newResult.TtlString = newResult.Ttl.String()

// 			// l.Trace().Time("expiresAt", expiresAt).Msg("trace log for record expiration")

// 			err := item.Value(func(v []byte) error {
// 				newResult.Id = string(k)
// 				newResult.Path = string(v)
// 				return nil
// 			})
// 			if err != nil {
// 				return err
// 			}

// 			// Formulate external URL
// 			newResult.Url = fmt.Sprintf("%s/links/%s", viper.GetString("server.externalurl"), newResult.Id)

// 			results = append(results, newResult)
// 		}
// 		return nil
// 	})
// 	return results, err
// }

// New Seclink DB
func NewSeclinkDb() ISeclinkDb {
	return &SSeclinkDb{}
}
