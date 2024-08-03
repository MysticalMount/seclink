package api

import (
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"seclink/db"
	"seclink/log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/mazen160/go-random"
	"github.com/spf13/viper"
)

//go:embed resources/*
var res embed.FS

type SCreateLink struct {
	Filepath  string        `json:"path"`
	TtlString string        `json:"ttl"`
	Ttl       time.Duration `json:"-"`
}

type ISeclinkApi interface {
	Start() error
}

type SSeclinkApi struct {
	db db.ISeclinkDb
}

// Starts the api server
func (a *SSeclinkApi) Start() error {

	// Prepare HTML template rendering system from embedded resources
	httpFS := http.FS(res)
	engine := html.NewFileSystem(httpFS, ".gohtml")

	// Public API and port
	app := fiber.New()
	app.Get("/get/:id", a.GetLink)

	// Private admin API and port
	admin := fiber.New(fiber.Config{Views: engine}) // Ensure we load the HTML template rendering engine
	admin.Use("/static", filesystem.New(filesystem.Config{
		Root:       httpFS,
		PathPrefix: "resources/static",
		Browse:     true,
	}))
	admin.Get("/", a.AdminUI)
	admin.Post("/links/share", a.CreateLink)

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
	var err error

	if err := c.BodyParser(&input); err != nil {
		l.Error().Err(err).Msg("Invalid input")
		return err
	}

	// Convert TTL string to time.Duration
	input.Ttl, err = time.ParseDuration(input.TtlString)
	if err != nil {
		l.Error().
			Err(err).
			Str("ttlstring", input.TtlString).
			Msg("Could not convert time string to duration")
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

		err = a.db.Set([]byte(id), []byte(input.Filepath), input.Ttl)

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

// Returns a list of relative filenames from the data directory, excludes db folder
func GetFileList() ([]SFile, error) {
	var files []SFile
	err := filepath.Walk(viper.GetString("server.datapath"), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			// Get relative path
			relPath, err := filepath.Rel(viper.GetString("server.datapath"), path)
			if err == nil {
				files = append(files, SFile{Path: relPath, TtlString: viper.GetDuration("links.defaultttl").String()})
			}
		}
		return nil
	})
	return files, err
}

// If link exists and has not expired then return downloaded file
func (a *SSeclinkApi) AdminUI(c *fiber.Ctx) error {
	l := log.Get()
	var err error

	l.Trace().Msg("Root page called")

	sharedLinks := []SSharedLink{{
		Path:      "hello.txt",
		Url:       viper.GetString("server.externalurl"),
		TtlString: "2h",
	}}

	files, err := GetFileList()
	if err != nil {
		l.Error().
			Err(err).
			Str("datapath", viper.GetString("server.datapath")).
			Msg("Could not list files in data path")
		return err
	}

	data := SUiData{
		SharedLinks: sharedLinks,
		Files:       files,
	}

	return c.Render("root", data)
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
