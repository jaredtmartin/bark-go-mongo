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
