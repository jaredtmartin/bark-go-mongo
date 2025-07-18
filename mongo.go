package bark

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Key string

const DbNameKey Key = "dbName"
const NowKey Key = "now"
const MockDbErrorKey Key = "mockDbError"

// var client *mongo.Client
var dbs = make(map[string]*mongo.Database)

// Connect to the MongoDB database
// If a database name is provided, it will connect to that database
// If no database name is provided, it will use the MONGO_DB environment variable
// If the database is already connected, it will return the existing connection
// If the connection fails, it will return an error
func Connect(db ...string) (*mongo.Database, error) {
	// fmt.Println("Connect db")
	uri := os.Getenv("MONGO_URI")
	env := os.Getenv("ENV")

	var dbName string
	if len(db) > 0 {
		dbName = db[0]
	} else {
		dbName = os.Getenv("MONGO_DB")
	}
	if dbs[dbName] == nil {
		if env != "test" {
			log.Printf("connecting to db: %s:%s\n", uri, dbName)
		}
		mc, err := mongo.Connect(options.Client().ApplyURI(uri))
		if err != nil {
			return nil, fmt.Errorf("error connecting to db: %v", err)
		}
		// client = mc
		dbs[dbName] = mc.Database(dbName)
	}
	return dbs[dbName], nil
}

// Get the database connection
// If the database name is not in the context, it will return an error
// If the database is not connected, it will connect to the database
// If the connection fails, it will return an error
// If the connection is successful, it will return the database connection
func Db(ctx context.Context) (*mongo.Database, error) {
	// fmt.Println("getting db")
	dbName, ok := ctx.Value(DbNameKey).(string)
	if !ok {
		return nil, fmt.Errorf("%s not found in context", DbNameKey)
	}
	mockError, ok := ctx.Value(MockDbErrorKey).(string)
	if ok {
		return nil, errors.New(mockError)
	}
	if dbs[dbName] == nil {
		return Connect(dbName)
	}
	return dbs[dbName], nil
}
