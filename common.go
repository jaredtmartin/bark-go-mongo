package bark

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrClearCanOnlyBeUsedOnDbsStartingWithTest = errors.New("to prevent accidents, clear method can only be used on databases whose names start with 'test'")

// Returns all documents matching the filter
func Find(collection *mongo.Collection, filter bson.M, results interface{}, opts *options.FindOptionsBuilder, ctx context.Context) error {
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return fmt.Errorf("error fetching documents: %v", err)
	}
	if err = cursor.All(ctx, results); err != nil {
		return fmt.Errorf("error decoding documents: %v", err)
	}
	return nil
}

// Returns the total number of documents matching the filter
func Count(collection *mongo.Collection, filter bson.M, ctx context.Context) (int64, error) {
	return collection.CountDocuments(ctx, filter)
}

// Returns the total number of documents matching the filter and returns the results
func FindAndCount(collection *mongo.Collection, filter bson.M, results interface{}, opts *options.FindOptionsBuilder, ctx context.Context) (int64, error) {
	err := Find(collection, filter, results, opts, ctx)
	if err != nil {
		return 0, err
	}
	count, err2 := Count(collection, filter, ctx)
	return count, err2
}

// Returns all documents in the collection
func All(collection *mongo.Collection, results interface{}, opts *options.FindOptionsBuilder, ctx context.Context) error {
	filter := bson.M{}
	return Find(collection, filter, results, opts, ctx)
}
