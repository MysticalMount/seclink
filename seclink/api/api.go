package api

import (
	"archive/zip"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"seclink/db"
	"seclink/log"
	"strings"
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

const mdName = "page.md"

//go:embed resources/*
var res embed.FS

type CreateLinkParams struct {
	PostName  string        `json:"post_name"`
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
	app.Use("/links", a.GetLink)

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
	admin.Post("/api/v1/links/share", a.CreateLink)
	// admin.Post("/api/v1/files/upload", a.UploadFile)
	admin.Post("/api/v1/pages/upload", a.UploadPage)

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

	// We need to parse the incoming route
	// and then use that as the ID for the link in the database

	splitPath := strings.Split(c.Path(), "/")
	if len(splitPath) < 3 {
		err := fmt.Errorf("missing ID")
		l.Error().
			Err(err).
			Msg("Missing ID field")
		return err
	}
	id := splitPath[2]

	// Validate that id exists in the path and is valid
	if len(id) != 64 || !regexp.MustCompile(`^([A-Za-z0-9]{64})$`).MatchString(id) {
		err := fmt.Errorf("ID %s is not a valid ID", id)
		l.Error().
			Err(err).
			Str("ID", id).
			Msg("Not a valid ID")
		return err
	}

	// See if the ID exists in the database
	getLinkRow, err := a.db.GetLink(id)
	if err != nil {
		l.Error().
			Err(err).
			Str("ID", id).
			Msg("Could not find id in database")
		return err
	}

	// Based on the Link.Expires unix epoch, determine if the link has expired or not
	if time.Now().Unix() > getLinkRow.Link.Expires {
		err = fmt.Errorf("%s has expired at %d", id, getLinkRow.Link.Expires)
		l.Error().
			Err(err).
			Str("ID", id).
			Msg("Link has expired")

		// As the link has expired, we want to use a.db.DeleteLink to remove the link
		err = a.db.DeleteLink(id)
		if err != nil {
			l.Error().
				Err(err).
				Str("ID", id).
				Msg("Could not be deleted from database")
			return err
		}

		return err
	}

	// Determine paths
	postPath := filepath.Join(viper.GetString("server.datapath"), getLinkRow.Post.Path)
	mdPath := filepath.Join(postPath, mdName)

	// If length of path segments is only links and the id, and nothing else has been specified then we only want to render the page.md
	// and serve the html for this file
	if len(splitPath) == 3 {
		// Check for page.md file
		exists, err := pathExists(mdPath)
		if err != nil || !exists {
			l.Error().
				Err(err).
				Str("ID", id).
				Msgf("Could not find %s in %s", mdName, postPath)
			return err
		}

		// Read in the markdown file
		md, err := os.ReadFile(mdPath)
		if err != nil {
			l.Error().
				Err(err).
				Str("ID", id).
				Str("mdPath", mdPath).
				Msg("Could not read in page")
			return err
		}

		// Parse the markdown file
		page, err := parseMarkdownFile(md)
		if err != nil {
			l.Error().
				Err(err).
				Str("mdPath", mdPath).
				Msg("Failed to parse markdown")
		}

		// Serve the HTML
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return c.Send([]byte(page.Content))
	}

	// TODO: Serve content if not root requested

	return nil
}

func (a *SSeclinkApi) CreateLink(c *fiber.Ctx) error {
	l := log.Get()

	var input CreateLinkParams
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

	absoluteFilePath := filepath.Join(viper.GetString("server.datapath"), input.PostName)
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

func (a *SSeclinkApi) UploadPage(c *fiber.Ctx) error {
	l := log.Get()

	l.Trace().Msg("UploadFile called")

	file, err := c.FormFile("binaryFile")
	if err != nil {
		l.Error().Err(err).Msg("Unable to load form data")
		return err
	}

	// Check is a zip file, required for page upload
	if filepath.Ext(file.Filename) == ".zip" {
		l.Info().Str("Filename", file.Filename).Msg("File extension zip is valid")
	} else {
		err := fmt.Errorf("must provide a zip file")
		l.Error().Err(err).Str("Filename", file.Filename).Msg("File extension zip is not valid")
		return err
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "page-upload")
	if err != nil {
		l.Error().Err(err).Str("Filename", file.Filename).Msg("Error creating temporary directory")
	} else {
		l.Info().Err(err).Str("Filename", file.Filename).Str("tempDir", tempDir).Msg("Successfully created temporary directory")
	}
	defer os.RemoveAll(tempDir)

	savePath := filepath.Join(tempDir, file.Filename)

	// Check for errors:
	if err == nil {
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

	// Extract the ZIP file, read markdown metadata
	extractPath := filepath.Join(tempDir, "extract")
	err = Unzip(savePath, extractPath)
	if err != nil {
		l.Error().Err(err).Str("Filename", file.Filename).Str("extractPath", extractPath).Str("zipFile", savePath).Msg("Error unzipping file")
		return err
	}

	l.Info().Str("Filename", file.Filename).Str("extractPath", extractPath).Str("zipFile", savePath).Msg("Successfull unzip")

	// Check the page.md file exists at a minimum this must exist
	mdPath := filepath.Join(extractPath, mdName)
	if exists, err := pathExists(mdPath); err != nil || !exists {
		l.Error().Err(err).Str("Filename", file.Filename).Str("mdPath", mdPath).Msgf("No %s file was found within zip", mdName)
	}

	l.Info().Err(err).Str("Filename", file.Filename).Str("mdPath", mdPath).Msgf("%s file was found, proceeding to process metadata", mdName)

	// Process metadata

	// Read in md file
	mdData, err := os.ReadFile(mdPath)
	if err != nil {
		l.Error().Err(err).Str("Filename", file.Filename).Str("mdPath", mdPath).Msgf("%s file could not be read", mdName)
	}

	// Parse the file
	Page, err := parseMarkdownFile(mdData)
	if err != nil {
		l.Error().Err(err).Str("Filename", file.Filename).Str("mdPath", mdPath).Msgf("%s file could not be parsed successfully, check format", mdName)
	}

	// Check that we have a slug name in the metadata, this will be the page name reference, folder name, and unique ID in the database
	if Page.Slug != "" {

		// Determine data path for page
		pagePath := filepath.Join(viper.GetString("server.datapath"), Page.Slug)

		// Re-unzip the page data in the page folder
		err := Unzip(savePath, pagePath)
		if err != nil {
			l.Error().Err(err).Str("Filename", file.Filename).Str("pagePath", pagePath).Msg("Failed to unzip to page data directory")
			return err
		}

		// Create an entry in the database for the page
		// TODO: Check there isnt an existing record and dont do anything if there already is/update path
		err = a.db.CreatePost(db.Post{Name: Page.Slug, Path: Page.Slug})
		if err != nil {
			l.Error().Err(err).Str("Filename", file.Filename).Str("pagePath", pagePath).Msg("Failed to create database record for page")
			return err
		}
	} else {
		err = fmt.Errorf("slug metadata must be provided and not blank")
		l.Error().Err(err).Str("Filename", file.Filename).Msg("Missing slug metadata")
		return err
	}

	// Return success
	err = c.SendString("Page upload successful!")
	if err != nil {
		return err
	}

	// Update post table
	data, err := a.GetUiData()
	if err != nil {
		l.Error().Err(err).Msg("failed to get required ui data")
		return err
	}

	return a.Render(c, AdminPostTable(data.Posts))
}

// Get all current data on the app, used for rendering UI pages
func (a *SSeclinkApi) GetUiData() (UiData, error) {

	l := log.Get()

	ctx := context.Background()

	links, err := a.db.Queries().GetAllLinks(ctx)
	if err != nil {
		l.Error().Err(err).Msg("failed to get links from db")
		return UiData{}, err
	}

	posts, err := a.db.Queries().GetAllPosts(ctx)
	if err != nil {
		l.Error().Err(err).Msg("failed to get posts from db")
		return UiData{}, err
	}

	return UiData{
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

// Helper functions

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

// Extract zip
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	err = os.MkdirAll(dest, 0700)
	if err != nil {
		return err
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
