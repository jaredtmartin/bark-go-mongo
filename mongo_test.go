package bark_test

import (
	"os"
	"testing"

	"github.com/jaredtmartin/bark-go-mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	// Mock environment variables
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "testdb")

	// Test case: Connect with default database from environment variable
	t.Run("Connect with default database", func(t *testing.T) {
		db, err := bark.Connect()
		require.NoError(t, err)
		assert.NotNil(t, db)
		assert.Equal(t, "testdb", db.Name())
	})

	// Test case: Connect with a specific database name
	t.Run("Connect with specific database", func(t *testing.T) {
		dbName := "specificdb"
		db, err := bark.Connect(dbName)
		require.NoError(t, err)
		assert.NotNil(t, db)
		assert.Equal(t, dbName, db.Name())
	})

}

func TestInvalidUri(t *testing.T) {
	os.Setenv("MONGO_URI", "invalid-uri")
	_, err := bark.Connect()
	assert.Error(t, err)
}
