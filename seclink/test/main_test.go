package test

import (
	"os"
	"testing"

	"seclink/api"
	"seclink/db"
	"seclink/log"
)

// TestMain runs before all tests in the package
func TestMain(m *testing.M) {
	// Set up any global test environment
	setupTestEnvironment()
	
	// Run the tests
	code := m.Run()
	
	// Clean up after tests
	teardownTestEnvironment()
	
	// Exit with the test status code
	os.Exit(code)
}

func setupTestEnvironment() {
	// Initialize logger with test level
	log.InitLog(0)
	
	// Set environment variables for testing
	os.Setenv("SECLINK_DATA_PATH", "./testdata")
	
	// Create test data directory if it doesn't exist
	os.MkdirAll("./testdata", 0755)
}

func teardownTestEnvironment() {
	// Clean up test data
	os.RemoveAll("./testdata")
}

// TestIntegration performs integration tests between components
func TestIntegration(t *testing.T) {
	// Create a test database
	testDb := db.NewSeclinkDb()
	err := testDb.Start()
	if err != nil {
		t.Fatalf("Failed to start test database: %v", err)
	}
	
	err = testDb.Migrate()
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	
	defer testDb.Close()
	
	// Create API with the real database
	secLinkApi := api.NewSeclinkApi(testDb)
	
	// Perform integration tests here
	// For example, create a link and verify it can be retrieved
	
	// Note: In a real test we'd use the API methods directly
	// but this gives a structure for writing integration tests
}