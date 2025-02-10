package api

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"seclink/db"
	"seclink/log"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mazen160/go-random"
	"github.com/spf13/viper"
)

//go:embed resources/*
var res embed.FS

type SCreateLink struct {
	PostName  string        `json:"path"`
	TtlString string        `json:"ttl"`
	Ttl       time.Duration `json:"-"`
}

type ISeclinkApi interface {
	Start() error
}

type SSeclinkApi struct {
	db            db.ISeclinkDb
	dataFilesPath string // The root data path is stored globally, but the files sub-folder is a constant, this stores that path so we dont have to repeatedly determine the sub-folder
}

// Starts the api server
func (a *SSeclinkApi) Start() error {

	l := log.Get()

	// Prepare HTML template rendering system from embedded resources
	httpFS := http.FS(res)

	// Public API and port
	app := fiber.New()
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &l,
	}))
	app.Use(recover.New())
	// app.Get("/links/:id", a.GetLink)

	// Private admin API and port
	// TODO: Make the BodyLimit in MB a configurable option
	admin := fiber.New(fiber.Config{BodyLimit: 2000 * 1024 * 1024}) // Ensure we load the HTML template rendering engine
	admin.Use("/static", filesystem.New(filesystem.Config{
		Root:       httpFS,
		PathPrefix: "resources/static",
		Browse:     true,
	}))
	admin.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &l,
	}))
	app.Use(recover.New())
	admin.Get("/admin", a.AdminUI)
	// admin.Post("/api/v1/links/share", a.CreateLink)
	admin.Post("/api/v1/files/upload", a.UploadFile)

	// Start admin port listening, as a goroutine
	go admin.Listen(fmt.Sprintf("0.0.0.0:%d", viper.GetInt("server.adminport")))

	// Start public port
	err := app.Listen(fmt.Sprintf("0.0.0.0:%d", viper.GetInt("server.port")))
	if err != nil {
		return err
	}

	return nil

}

// // If link exists and has not expired then return downloaded file
// func (a *SSeclinkApi) GetLink(c *fiber.Ctx) error {
// 	l := log.Get()

// 	if id := c.Params("id"); id != "" {

// 		// See if the ID exists in the database
// 		filePath, err := a.db.Get([]byte(id))
// 		if err != nil {
// 			l.Error().
// 				Err(err).
// 				Str("ID", id).
// 				Msg("Could not find id in database")
// 			return err
// 		}

// 		// Check the file exists
// 		absoluteFilePath := filepath.Join(a.dataFilesPath, string(filePath))
// 		exists, err := pathExists(absoluteFilePath)
// 		if err != nil {
// 			l.Error().
// 				Err(err).
// 				Str("ID", id).
// 				Str("AbsoluteFilePath", absoluteFilePath).
// 				Msg("Error occurred checking if file exists")
// 			return err
// 		}
// 		if !exists {
// 			l.Error().
// 				Str("ID", id).
// 				Str("AbsoluteFilePath", absoluteFilePath).
// 				Msg("File does not exist")
// 			return err
// 		}

// 		l.Info().Str("AbsoluteFilePath", absoluteFilePath).Str("ID", id).Msg("Downloading file")
// 		c.Download(absoluteFilePath, string(filePath))

// 	} else {
// 		l.Error().
// 			Msg("An empty id was provided on the route")
// 		return fmt.Errorf("an empty id was provided on the route")
// 	}

// 	return nil
// }

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

	post, err := a.db.GetPost(input.PostName)
	if err != nil {
		l.Error().Err(err).Str("PostName", input.PostName).Msg("Post does not exist in DB or error finding record")
		return err
	} else {
		l.Info().Str("PostName", post.Name).Msg("Found Post in DB")
	}

	absoluteFilePath := filepath.Join(a.dataFilesPath, post.Path)
	exists, err := pathExists(absoluteFilePath)
	if err != nil {
		l.Error().Err(err).Str("FilePath", absoluteFilePath).Msg("Post file path does not exist")
		return err
	}

	if exists {
		id, err := GenerateLink()
		if err != nil {
			l.Error().Err(err).Str("PostName", post.Name).Str("id", id).Msg("An error occurred generating a random ID")
			return err
		}
		l.Info().Str("id", id).Msg("Generated ID")

		// Formulate the expires time
		expiresAt := time.Now().Local().Add(input.Ttl)
		expiresAtUnixEpoch := expiresAt.Unix()

		link := db.Link{ID: id, Expires: expiresAtUnixEpoch, PostName: post.Name}

		err = a.db.CreateLink(link)

		if err != nil {
			l.Error().Err(err).Interface("link", link).Msg("An error occurred creating the link record in the database")
			return err
		}

	} else {
		l.Error().Err(err).Str("PostName", post.Name).Str("AbsoluteFilePath", absoluteFilePath).Msg("Path for post does not exist")
		return fmt.Errorf("post path does not exist")
	}

	data, err := a.GetUiData()
	if err != nil {
		l.Error().Err(err).Msg("failed to get required ui data")
		return err
	}

	return a.Render(c, AdminLinksTable(data.Links))

}

func GenerateLink() (string, error) {
	data, err := random.String(64)
	return data, err
}

// Returns a list of relative filenames from the data directory, excludes db folder
func (a *SSeclinkApi) GetFileList() ([]SFile, error) {
	var files []SFile
	err := filepath.Walk(a.dataFilesPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			// Get relative path
			relPath, err := filepath.Rel(a.dataFilesPath, path)
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

	data, err := a.GetUiData()
	if err != nil {
		l.Error().Err(err).Msg("failed to get required ui data")
		return err
	}

	return a.Render(c, AdminUiPage(data.Links, data.Posts))
}

func (a *SSeclinkApi) UploadFile(c *fiber.Ctx) error {
	l := log.Get()

	l.Trace().Msg("UploadFile called")

	file, err := c.FormFile("binaryFile")

	// Check for errors:
	if err == nil {
		savePath := filepath.Join(a.dataFilesPath, file.Filename)
		l.Info().
			Str("savePath", savePath).
			Str("Filename", file.Filename).
			Msg("file upload successful, saving file")
		// 👷 Save file to root directory:
		err = c.SaveFile(file, savePath)
		if err != nil {
			l.Error().
				Err(err).
				Str("savePath", savePath).
				Str("Filename", file.Filename).
				Msg("failed to save file to the save path")
			return err
		}
	} else {
		l.Error().
			Err(err).
			Str("Filename", file.Filename).
			Msg("failed to upload file")
		return err
	}

	err = c.SendString("File upload successful!")
	if err != nil {
		return err
	}

	data, err := a.GetUiData()
	if err != nil {
		l.Error().Err(err).Msg("failed to get required ui data")
		return err
	}

	return a.Render(c, AdminPostTable(data.Posts))
}

// Get all current data on the app, used for rendering UI pages
func (a *SSeclinkApi) GetUiData() (SUiData, error) {

	l := log.Get()

	ctx := context.Background()

	links, err := a.db.Queries().GetAllLinks(ctx)
	if err != nil {
		l.Error().Err(err).Msg("failed to get links from db")
		return SUiData{}, err
	}

	posts, err := a.db.Queries().GetAllPosts(ctx)
	if err != nil {
		l.Error().Err(err).Msg("failed to get links from db")
		return SUiData{}, err
	}

	if err != nil {

		return SUiData{}, err
	}

	return SUiData{
		Links: links,
		Posts: posts,
	}, nil

}

func (a *SSeclinkApi) Render(c *fiber.Ctx, component templ.Component, options ...func(*templ.ComponentHandler)) error {
	componentHandler := templ.Handler(component)
	for _, o := range options {
		o(componentHandler)
	}
	return adaptor.HTTPHandler(componentHandler)(c)
}

// New Seclink API
func NewSeclinkApi(db db.ISeclinkDb) ISeclinkApi {
	return &SSeclinkApi{
		db:            db,
		dataFilesPath: filepath.Join(viper.GetString("server.datapath"), "files"),
	}
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
