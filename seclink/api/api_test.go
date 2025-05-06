package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"seclink/db"
)

// MockSeclinkDb is a mock implementation of db.ISeclinkDb
type MockSeclinkDb struct {
	mock.Mock
}

func (m *MockSeclinkDb) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSeclinkDb) Migrate() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSeclinkDb) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSeclinkDb) Queries() *db.Queries {
	args := m.Called()
	return args.Get(0).(*db.Queries)
}

func (m *MockSeclinkDb) GetAllLinks() ([]db.GetAllLinksRow, error) {
	args := m.Called()
	return args.Get(0).([]db.GetAllLinksRow), args.Error(1)
}

func (m *MockSeclinkDb) CreateLink(link db.Link) error {
	args := m.Called(link)
	return args.Error(0)
}

func (m *MockSeclinkDb) CreatePost(post db.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *MockSeclinkDb) DeletePost(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockSeclinkDb) DeleteLink(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockSeclinkDb) GetLink(id string) (db.GetLinkRow, error) {
	args := m.Called(id)
	return args.Get(0).(db.GetLinkRow), args.Error(1)
}

func (m *MockSeclinkDb) GetAllPosts() ([]db.Post, error) {
	args := m.Called()
	return args.Get(0).([]db.Post), args.Error(1)
}

func (m *MockSeclinkDb) GetPost(name string) (db.Post, error) {
	args := m.Called(name)
	return args.Get(0).(db.Post), args.Error(1)
}

// Setup test app with mocked database
func setupTestApp() (*fiber.App, *MockSeclinkDb, *SSeclinkApi) {
	mockDb := new(MockSeclinkDb)
	api := &SSeclinkApi{
		db:            mockDb,
		dataFilesPath: "./testdata", // Use a test directory
	}
	app := fiber.New()
	
	// Register routes similar to the actual application
	app.Get("/link/:id", api.GetLink)
	app.Post("/link", api.CreateLink)
	
	return app, mockDb, api
}

// Helper function to make test HTTP requests
func makeRequest(app *fiber.App, method, path string, body io.Reader) (*http.Response, error) {
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	return app.Test(req)
}

// TestCreateLink tests the CreateLink endpoint
func TestCreateLink(t *testing.T) {
	app, mockDb, _ := setupTestApp()
	
	// Test case: successful link creation
	t.Run("Successful link creation", func(t *testing.T) {
		// Setup mock expectations
		mockDb.On("CreateLink", mock.AnythingOfType("db.Link")).Return(nil)
		
		// Create request body
		requestBody := `{
			"post_name": "test-post",
			"ttl": "24h"
		}`
		
		// Make request
		resp, err := makeRequest(app, "POST", "/link", strings.NewReader(requestBody))
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		
		// Verify mock was called
		mockDb.AssertExpectations(t)
	})
	
	// Test case: invalid TTL format
	t.Run("Invalid TTL format", func(t *testing.T) {
		// Create request body with invalid TTL
		requestBody := `{
			"post_name": "test-post",
			"ttl": "invalid-ttl"
		}`
		
		// Make request
		resp, err := makeRequest(app, "POST", "/link", strings.NewReader(requestBody))
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
	
	// Test case: malformed JSON input
	t.Run("Malformed input", func(t *testing.T) {
		// Create invalid JSON
		requestBody := `{
			"post_name": "test-post",
			"ttl": "24h",
		}` // Note the trailing comma
		
		// Make request
		resp, err := makeRequest(app, "POST", "/link", strings.NewReader(requestBody))
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// TestGetLink tests the GetLink endpoint
func TestGetLink(t *testing.T) {
	app, mockDb, _ := setupTestApp()
	
	// Test case: successful link retrieval
	t.Run("Successful link retrieval", func(t *testing.T) {
		// Setup mock expectations
		mockLink := db.GetLinkRow{
			ID:        "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			PostName:  "test-post",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		
		mockDb.On("GetLink", mockLink.ID).Return(mockLink, nil)
		
		// Make request
		resp, err := makeRequest(app, "GET", "/link/"+mockLink.ID, nil)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		
		// Verify mock was called
		mockDb.AssertExpectations(t)
	})
	
	// Test case: invalid ID format
	t.Run("Invalid ID format", func(t *testing.T) {
		// Make request with invalid ID
		resp, err := makeRequest(app, "GET", "/link/invalid-id", nil)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
	
	// Test case: link not found
	t.Run("Link not found", func(t *testing.T) {
		// Valid ID but doesn't exist
		validID := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		
		// Setup mock expectations to return error
		mockDb.On("GetLink", validID).Return(db.GetLinkRow{}, db.ErrNotFound)
		
		// Make request
		resp, err := makeRequest(app, "GET", "/link/"+validID, nil)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		
		// Verify mock was called
		mockDb.AssertExpectations(t)
	})
}