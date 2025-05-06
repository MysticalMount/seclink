package test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	
	"seclink/api"
	"seclink/db"
)

// Helper function to create a test app with real database connection
func setupTestAppWithRealDb(t *testing.T) (*fiber.App, db.ISeclinkDb) {
	// Set up test database path
	testDbPath := filepath.Join(t.TempDir(), "testdb")
	os.Setenv("SECLINK_DATA_PATH", testDbPath)
	os.MkdirAll(testDbPath, 0755)
	
	// Create database
	testDb := db.NewSeclinkDb()
	err := testDb.Start()
	assert.NoError(t, err, "Database should start without error")
	
	err = testDb.Migrate()
	assert.NoError(t, err, "Database migration should succeed")
	
	// Create API with real database
	seclinkApi := api.NewSeclinkApi(testDb).(*api.SSeclinkApi)
	
	// Create test app
	app := fiber.New()
	
	// Register routes
	app.Get("/link/:id", seclinkApi.GetLink)
	app.Post("/link", seclinkApi.CreateLink)
	app.Get("/admin", seclinkApi.AdminUI)
	app.Post("/upload", seclinkApi.UploadFile)
	app.Post("/page", seclinkApi.UploadPage)
	
	return app, testDb
}

// Helper function to create a file upload request for testing
func createFileUploadRequest(t *testing.T, url, fieldName, fileName string, fileContent []byte, extraFields map[string]string) *http.Request {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	
	// Add file part
	part, err := writer.CreateFormFile(fieldName, fileName)
	assert.NoError(t, err)
	
	_, err = part.Write(fileContent)
	assert.NoError(t, err)
	
	// Add extra form fields
	for key, value := range extraFields {
		err = writer.WriteField(key, value)
		assert.NoError(t, err)
	}
	
	err = writer.Close()
	assert.NoError(t, err)
	
	// Create request
	req, err := http.NewRequest("POST", url, body)
	assert.NoError(t, err)
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// Helper function to make JSON request for testing
func createJSONRequest(t *testing.T, method, url string, jsonBody []byte) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	
	req.Header.Set("Content-Type", "application/json")
	return req
}

// Helper function to read response body
func readResponseBody(t *testing.T, resp *http.Response) []byte {
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	return body
}