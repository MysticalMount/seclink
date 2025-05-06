package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestAdminUI tests the admin UI endpoint
func TestAdminUI(t *testing.T) {
	app, mockDb, api := setupTestApp()
	
	// Register the admin UI route for testing
	app.Get("/admin", api.AdminUI)
	
	t.Run("Access admin UI", func(t *testing.T) {
		// Setup mock expectations
		mockLinks := []db.GetAllLinksRow{
			{
				ID:        "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				PostName:  "test-post",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
		}
		
		mockPosts := []db.Post{
			{
				Name:    "test-post",
				Content: []byte("Test post content"),
			},
		}
		
		mockFiles := []SFile{
			{
				Name: "test.txt",
				Path: "/files/test.txt",
			},
		}
		
		mockDb.On("GetAllLinks").Return(mockLinks, nil)
		mockDb.On("GetAllPosts").Return(mockPosts, nil)
		
		// Make the request
		resp, err := makeRequest(app, "GET", "/admin", nil)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		
		// Verify mock was called
		mockDb.AssertExpectations(t)
	})
}

// TestGetUiData tests the GetUiData function that provides data for the admin UI
func TestGetUiData(t *testing.T) {
	_, mockDb, api := setupTestApp()
	
	// Setup mock expectations
	mockLinks := []db.GetAllLinksRow{
		{
			ID:        "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			PostName:  "test-post",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}
	
	mockPosts := []db.Post{
		{
			Name:    "test-post",
			Content: []byte("Test post content"),
		},
	}
	
	mockDb.On("GetAllLinks").Return(mockLinks, nil)
	mockDb.On("GetAllPosts").Return(mockPosts, nil)
	
	// Call the function
	uiData, err := api.GetUiData()
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(uiData.Links), "Should have 1 link")
	assert.Equal(t, 1, len(uiData.Posts), "Should have 1 post")
	
	// Verify mock was called
	mockDb.AssertExpectations(t)
}