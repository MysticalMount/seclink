package db

import (
	"context"
	"database/sql"
	"embed"
	"path/filepath"
	"seclink/log"

	// _ "github.com/glebarez/go-sqlite"

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

	// Control functions
	Start() error
	Migrate() error
	Close() error
	Queries() *Queries

	// Repository functions
	GetAllLinks() ([]GetAllLinksRow, error)
	CreateLink(link Link) error
	CreatePost(post Post) error
	DeletePost(name string) error
	DeleteLink(id string) error
	GetLink(id string) (GetLinkRow, error)
	GetAllPosts() ([]Post, error)
}

type SSeclinkDb struct {
	db      *sql.DB
	queries *Queries
}

// Control functions

func (d *SSeclinkDb) Start() error {

	l := log.Get()

	dbPath := filepath.Join(viper.GetString("server.datapath"), databaseFilename)
	l.Info().
		Str("dbPath", dbPath).
		Msg("Attempting to open Sqlite database")

	// connect
	db, err := sql.Open("sqlite", dbPath)
	if err != nil { // list all authors

		l.Error().
			Err(err).
			Msg("An error was encountered opening the Sqlite database")
		return err
	}

	// Successfully connected
	l.Info().
		Msg("Sqlite connection established")

	// Persist connection objects
	d.db = db
	d.queries = New(db)

	return nil
}

func (d *SSeclinkDb) Migrate() error {
	// NOTE: golang-migrate and sqlite3
	//
	// golang-migrate doesnt natively work with sqlite3 - as seclink has minimal data requirements in terms of performance, to avoid
	// these complications, seclink only uses sqlite (2)
	// In the future the work to use sqlite3 could be considered

	l := log.Get()

	// Create golang-migrate database driver
	driver, err := sqlite.WithInstance(d.db, &sqlite.Config{DatabaseName: databaseName})
	if err != nil {
		l.Error().
			Err(err).
			Msg("An error was encountered during database migration")
		return err
	}

	// Initialise source from embedded migrations
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		l.Error().
			Err(err).
			Msg("An error was encountered during database migration")
		return err
	}

	// Create new golang-migrate instance
	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		l.Error().
			Err(err).
			Msg("An error was encountered during database migration")
		return err
	}

	// Migrate all
	m.Up()

	// Indicate successful migration
	l.Info().
		Err(err).
		Msg("Successful database migrations")

	return err
}

// Closes the DB
func (d *SSeclinkDb) Close() error {
	return d.db.Close()
}

func (d *SSeclinkDb) Queries() *Queries {
	return d.queries
}

// Repository functions

// Gets all links in the db
func (d *SSeclinkDb) GetAllLinks() ([]GetAllLinksRow, error) {

	ctx := context.Background()

	links, err := d.queries.GetAllLinks(ctx)
	if err != nil {
		return nil, err
	}

	return links, err

}

// Add a link
func (d *SSeclinkDb) CreateLink(link Link) error {

	ctx := context.Background()

	err := d.queries.CreateLink(ctx, CreateLinkParams(link))
	if err != nil {
		return err
	}

	return nil

}

// Create post
func (d *SSeclinkDb) CreatePost(post Post) error {

	ctx := context.Background()

	err := d.queries.CreatePost(ctx, CreatePostParams(post))
	if err != nil {
		return err
	}

	return nil

}

// Delete post
func (d *SSeclinkDb) DeletePost(name string) error {

	ctx := context.Background()

	err := d.queries.DeletePost(ctx, name)
	if err != nil {
		return err
	}

	return nil

}

// Delete link
func (d *SSeclinkDb) DeleteLink(id string) error {

	ctx := context.Background()

	err := d.queries.DeleteLink(ctx, id)
	if err != nil {
		return err
	}

	return nil

}

// Get link
func (d *SSeclinkDb) GetLink(id string) (GetLinkRow, error) {

	ctx := context.Background()

	getLinkRow, err := d.queries.GetLink(ctx, id)
	if err != nil {
		return GetLinkRow{}, err
	}

	return getLinkRow, nil

}

// Get all posts
func (d *SSeclinkDb) GetAllPosts() ([]Post, error) {

	ctx := context.Background()

	allPosts, err := d.queries.GetAllPosts(ctx)
	if err != nil {
		return nil, err
	}

	return allPosts, nil

}

// New Seclink DB
func NewSeclinkDb() ISeclinkDb {
	return &SSeclinkDb{}
}
