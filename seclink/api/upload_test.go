package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestUploadFile tests the file upload endpoint
func TestUploadFile(t *testing.T) {
	app, mockDb, api := setupTestApp()
	
	// Register the upload route for testing
	app.Post("/upload", api.UploadFile)
	
	// Create a test directory for file uploads
	testDataDir := "./testdata"
	err := os.MkdirAll(testDataDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(testDataDir)
	
	t.Run("Successful file upload", func(t *testing.T) {
		// Create a test file
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "test.txt")
		assert.NoError(t, err)
		
		fileContent := []byte("This is a test file")
		_, err = part.Write(fileContent)
		assert.NoError(t, err)
		
		err = writer.Close()
		assert.NoError(t, err)
		
		// Create a request
		req, err := http.NewRequest("POST", "/upload", body)
		assert.NoError(t, err)
		
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		// Make the request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		
		// Verify the file was saved (would check but the path is determined by the API implementation)
	})
	
	t.Run("No file provided", func(t *testing.T) {
		// Create a request without a file
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		err = writer.Close()
		assert.NoError(t, err)
		
		req, err := http.NewRequest("POST", "/upload", body)
		assert.NoError(t, err)
		
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		// Make the request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// TestUploadPage tests the page upload endpoint
func TestUploadPage(t *testing.T) {
	app, mockDb, api := setupTestApp()
	
	// Register the upload route for testing
	app.Post("/page", api.UploadPage)
	
	// Create a test directory for page uploads
	testDataDir := "./testdata"
	err := os.MkdirAll(testDataDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(testDataDir)
	
	t.Run("Successful page upload", func(t *testing.T) {
		// Create a test file
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		
		// Add page name field
		err := writer.WriteField("name", "test-page")
		assert.NoError(t, err)
		
		// Add file field
		part, err := writer.CreateFormFile("file", "page.md")
		assert.NoError(t, err)
		
		fileContent := []byte("# Test Page\n\nThis is a test markdown page")
		_, err = part.Write(fileContent)
		assert.NoError(t, err)
		
		err = writer.Close()
		assert.NoError(t, err)
		
		// Setup mock expectations for database
		mockPost := mock.AnythingOfType("db.Post")
		mockDb.On("CreatePost", mockPost).Return(nil)
		
		// Create a request
		req, err := http.NewRequest("POST", "/page", body)
		assert.NoError(t, err)
		
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		// Make the request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		
		// Verify mock was called
		mockDb.AssertExpectations(t)
	})
	
	t.Run("Missing page name", func(t *testing.T) {
		// Create a test file without page name
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		
		// Add file field without name field
		part, err := writer.CreateFormFile("file", "page.md")
		assert.NoError(t, err)
		
		fileContent := []byte("# Test Page\n\nThis is a test markdown page")
		_, err = part.Write(fileContent)
		assert.NoError(t, err)
		
		err = writer.Close()
		assert.NoError(t, err)
		
		// Create a request
		req, err := http.NewRequest("POST", "/page", body)
		assert.NoError(t, err)
		
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		// Make the request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
	
	t.Run("Missing file", func(t *testing.T) {
		// Create a test request without file
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		
		// Add page name field
		err := writer.WriteField("name", "test-page")
		assert.NoError(t, err)
		
		err = writer.Close()
		assert.NoError(t, err)
		
		// Create a request
		req, err := http.NewRequest("POST", "/page", body)
		assert.NoError(t, err)
		
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		// Make the request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}