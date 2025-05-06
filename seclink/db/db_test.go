package db

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDatabaseOperations tests database operations using a test database
func TestDatabaseOperations(t *testing.T) {
	// Create a test database
	testDbPath := "./testdb"
	os.Setenv("SECLINK_DATA_PATH", testDbPath)
	defer os.RemoveAll(testDbPath)
	
	// Ensure test directory exists
	os.MkdirAll(testDbPath, 0755)
	
	// Initialize the database
	db := NewSeclinkDb()
	err := db.Start()
	assert.NoError(t, err, "Database should start without error")
	
	err = db.Migrate()
	assert.NoError(t, err, "Database migration should succeed")
	
	defer db.Close()
	
	// Test creating a link
	t.Run("Create and retrieve link", func(t *testing.T) {
		link := Link{
			ID:        "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			PostName:  "test-post",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		
		// Create link
		err := db.CreateLink(link)
		assert.NoError(t, err, "Should create link without error")
		
		// Retrieve link
		retrievedLink, err := db.GetLink(link.ID)
		assert.NoError(t, err, "Should retrieve link without error")
		assert.Equal(t, link.ID, retrievedLink.ID, "Retrieved link ID should match")
		assert.Equal(t, link.PostName, retrievedLink.PostName, "Retrieved link post name should match")
	})
	
	// Test creating a post
	t.Run("Create and retrieve post", func(t *testing.T) {
		post := Post{
			Name:    "test-post",
			Content: []byte("Test post content"),
		}
		
		// Create post
		err := db.CreatePost(post)
		assert.NoError(t, err, "Should create post without error")
		
		// Retrieve post
		retrievedPost, err := db.GetPost(post.Name)
		assert.NoError(t, err, "Should retrieve post without error")
		assert.Equal(t, post.Name, retrievedPost.Name, "Retrieved post name should match")
		assert.Equal(t, post.Content, retrievedPost.Content, "Retrieved post content should match")
	})
	
	// Test retrieving all links
	t.Run("Get all links", func(t *testing.T) {
		links, err := db.GetAllLinks()
		assert.NoError(t, err, "Should retrieve all links without error")
		assert.GreaterOrEqual(t, len(links), 1, "Should have at least one link")
	})
	
	// Test retrieving all posts
	t.Run("Get all posts", func(t *testing.T) {
		posts, err := db.GetAllPosts()
		assert.NoError(t, err, "Should retrieve all posts without error")
		assert.GreaterOrEqual(t, len(posts), 1, "Should have at least one post")
	})
	
	// Test deleting a link
	t.Run("Delete link", func(t *testing.T) {
		linkID := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		
		// Delete link
		err := db.DeleteLink(linkID)
		assert.NoError(t, err, "Should delete link without error")
		
		// Try to retrieve deleted link
		_, err = db.GetLink(linkID)
		assert.Error(t, err, "Should not find deleted link")
	})
	
	// Test deleting a post
	t.Run("Delete post", func(t *testing.T) {
		postName := "test-post"
		
		// Delete post
		err := db.DeletePost(postName)
		assert.NoError(t, err, "Should delete post without error")
		
		// Try to retrieve deleted post
		_, err = db.GetPost(postName)
		assert.Error(t, err, "Should not find deleted post")
	})
}