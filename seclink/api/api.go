package api

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"seclink/db"
	"seclink/log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mazen160/go-random"
	"github.com/spf13/viper"
)

type SCreateLink struct {
	Filepath string `json:"filepath"`
	Ttl      int32  `json:"ttl,omitempty"`
}

type ISeclinkApi interface {
	Start() error
}

type SSeclinkApi struct {
	db db.ISeclinkDb
}

// Starts the api server
func (a *SSeclinkApi) Start() error {

	// Public API and port
	app := fiber.New()
	app.Get("/get/:id", a.GetLink)

	// Private admin API and port
	admin := fiber.New()
	admin.Post("/link", a.CreateLink)

	// Start admin port listening, as a goroutine
	go admin.Listen(fmt.Sprintf("0.0.0.0:%d", viper.GetInt("server.adminport")))

	// Start public port
	err := app.Listen(fmt.Sprintf("0.0.0.0:%d", viper.GetInt("server.port")))
	if err != nil {
		return err
	}

	return nil

}

// If link exists and has not expired then return downloaded file
func (a *SSeclinkApi) GetLink(c *fiber.Ctx) error {
	l := log.Get()

	if id := c.Params("id"); id != "" {

		// See if the ID exists in the database
		filePath, err := a.db.Get([]byte(id))
		if err != nil {
			l.Error().
				Err(err).
				Str("ID", id).
				Msg("Could not find id in database")
			return err
		}

		// Check the file exists
		absoluteFilePath := filepath.Join(viper.GetString("server.datapath"), string(filePath))
		exists, err := pathExists(absoluteFilePath)
		if err != nil {
			l.Error().
				Err(err).
				Str("ID", id).
				Str("AbsoluteFilePath", absoluteFilePath).
				Msg("Error occurred checking if file exists")
			return err
		}
		if !exists {
			l.Error().
				Str("ID", id).
				Str("AbsoluteFilePath", absoluteFilePath).
				Msg("File does not exist")
			return err
		}

		l.Info().Str("AbsoluteFilePath", absoluteFilePath).Str("ID", id).Msg("Downloading file")
		c.Download(absoluteFilePath, string(filePath))

	} else {
		l.Error().
			Msg("An empty id was provided on the route")
		return fmt.Errorf("an empty id was provided on the route")
	}

	return nil
}

func (a *SSeclinkApi) CreateLink(c *fiber.Ctx) error {
	l := log.Get()

	var input SCreateLink

	if err := c.BodyParser(&input); err != nil {
		l.Error().Err(err).Msg("Invalid input")
		return err
	}

	l.Trace().Interface("input", input).Msg("Input")

	absoluteFilePath := filepath.Join(viper.GetString("server.datapath"), input.Filepath)
	exists, err := pathExists(absoluteFilePath)
	if err != nil {
		l.Error().Err(err).Str("FilePath", input.Filepath).Msg("An error occurred determining if filepath exists")
		return err
	}

	if exists {
		id, err := GenerateLink()
		if err != nil {
			l.Error().Err(err).Str("FilePath", input.Filepath).Str("ID", id).Msg("An error occurred generating a random ID")
			return err
		}
		l.Info().Str("id", id).Msg("Generated ID")

		err = a.db.Set([]byte(id), []byte(input.Filepath), time.Duration(input.Ttl)*time.Minute)

		if err != nil {
			l.Error().Err(err).Str("FilePath", input.Filepath).Str("ID", id).Msg("An error occurred inserting a record")
			return err
		}

	} else {
		l.Error().Err(err).Str("FilePath", input.Filepath).Str("AbsoluteFilePath", absoluteFilePath).Msg("Filepath does not exist")
		return fmt.Errorf("file does not exist")
	}

	return nil

}

func GenerateLink() (string, error) {
	data, err := random.String(64)
	return data, err
}

// New Seclink API
func NewSeclinkApi(db db.ISeclinkDb) ISeclinkApi {
	return &SSeclinkApi{db: db}
}

// Path exists
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
