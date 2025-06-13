package bark

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// A model that can be used to save to the database
type ModelWithCollection interface {
	SetCollectionName(name string)
	GetCollectionName() string
}

// A collection of models
type Collection[T ModelWithCollection] struct {
	Name       string
	collection *mongo.Collection
}

// Creates a new collection
func NewCollection[T ModelWithCollection](name string) *Collection[T] {
	return &Collection[T]{Name: name}
}

// Returns the mongo collection
func (c *Collection[T]) MongoCollection(ctx context.Context) (*mongo.Collection, error) {
	mockError, ok := ctx.Value(MockDbErrorKey).(string)
	if ok {
		return nil, errors.New(mockError)
	}
	if c.collection == nil {
		if c.Name == "" {
			return nil, fmt.Errorf("collection name is required")
		}
		db, err := Db(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get database: %v", err)
		}
		c.collection = db.Collection(c.Name)
	}
	return c.collection, nil
}

// Finds all documents matching the filter and returns a slice of T
func (c *Collection[T]) Find(filter bson.M, opts *options.FindOptionsBuilder, ctx context.Context) ([]T, error) {
	var results []T
	collection, err := c.MongoCollection(ctx)
	if err != nil {
		return results, fmt.Errorf("failed to get collection to save model to: %v", err)
	}
	// fmt.Println("collection.Name(): ", collection.Name())
	// fmt.Println("collection.Database().Name(): ", collection.Database().Name())
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return results, nil
	}
	if err != nil {
		return results, fmt.Errorf("error fetching documents: %v", err)
	}
	if err = cursor.All(ctx, &results); err != nil {
		return results, fmt.Errorf("error decoding documents: %v", err)
	}
	for i := range results {
		results[i].SetCollectionName(c.Name)
	}
	return results, nil
}

// Finds a single document matching the filter
func (c *Collection[T]) FindOne(filter bson.M, ctx context.Context) (T, error) {
	collection, err := c.MongoCollection(ctx)
	if err != nil {
		return *new(T), fmt.Errorf("failed to get collection to save model to: %v", err)
	}
	obj := *new(T)
	err = collection.FindOne(ctx, filter).Decode(&obj)
	if err == mongo.ErrNoDocuments {
		return *new(T), ErrNotFound
	}
	if err != nil {
		return *new(T), fmt.Errorf("error fetching documents: %v", err)
	}
	obj.SetCollectionName(c.Name)
	// fmt.Println("obj.Name(): ", obj.Name())
	return obj, nil
}

// Returns all documents in the collection
func (c *Collection[T]) All(ctx context.Context) ([]T, error) {
	return c.Find(bson.M{}, nil, ctx)
}

// Returns the number of documents matching the filter
func (c *Collection[T]) Count(filter bson.M, ctx context.Context) (int64, error) {
	collection, err := c.MongoCollection(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get collection to count: %v", err)
	}
	return collection.CountDocuments(ctx, filter)
}

// Returns and counts all documents matching the filter
func (c *Collection[T]) FindAndCount(filter bson.M, opts *options.FindOptionsBuilder, ctx context.Context) ([]T, int64, error) {
	collection, err := c.MongoCollection(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get collection to find and count: %v", err)
	}
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching documents: %v", err)
	}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting documents: %v", err)
	}
	var results []T
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, fmt.Errorf("error decoding documents: %v", err)
	}
	return results, count, nil
}

// Gets a single document with matching id
func (c *Collection[T]) Get(id string, ctx context.Context) (T, error) {
	filter := bson.M{"Id": id}
	return c.FindOne(filter, ctx)
}

// Deletes a single document matching the filter
func (c *Collection[T]) DeleteOne(filter bson.M, ctx context.Context) (*Result, error) {
	collection, err := c.MongoCollection(ctx)
	if err != nil {
		return EmptyResult(), fmt.Errorf("error getting collection to clear: %v", err)
	}
	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return ResultFromDelete(res), fmt.Errorf("error deleting documents: %v", err)
	}
	return ResultFromDelete(res), nil
}

// Deletes all documents matching the filter
func (c *Collection[T]) DeleteMany(filter bson.M, ctx context.Context) (*Result, error) {
	collection, err := c.MongoCollection(ctx)
	if err != nil {
		return EmptyResult(), fmt.Errorf("error getting collection to clear: %v", err)
	}
	res, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return ResultFromDelete(res), fmt.Errorf("error deleting documents: %v", err)
	}
	return ResultFromDelete(res), nil
}
