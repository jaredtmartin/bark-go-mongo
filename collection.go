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

type Collection[model any] struct {
	collection *mongo.Collection
}

func NewCollection[model any](name string, ctx context.Context) (*Collection[model], error) {
	database, err := Db(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting db for collection %s: %v", name, err)
	}
	collection := database.Collection(name)
	return &Collection[model]{
		collection: collection,
	}, nil
}

// // Returns a single document with the given id
func (c *Collection[model]) Get(id string, ctx context.Context) (*model, error) {
	filter := bson.M{"_id": id}
	return c.FindOne(filter, ctx)
}

func (c *Collection[model]) New() *model {
	m := new(model)
	// m.SetCollection(c.collection)
	return m
}

// Returns one document matching the filter
func (c *Collection[model]) FindOne(filter bson.M, ctx context.Context) (*model, error) {
	obj := new(model)
	err := c.collection.FindOne(ctx, filter).Decode(obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Returns one document matching the filter
func (c *Collection[model]) Clear(ctx context.Context) error {
	if c.collection.Database().Name()[:4] != "test" {
		return ErrClearCanOnlyBeUsedOnDbsStartingWithTest
	}
	filter := bson.M{}
	_, err := c.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("error clearing collection: %v", err)
	}
	return nil
}

// Returns all documents matching the filter
func (c *Collection[model]) Find(filter bson.M, results interface{}, opts *options.FindOptionsBuilder, ctx context.Context) error {
	cursor, err := c.collection.Find(ctx, filter, opts)
	if err != nil {
		return fmt.Errorf("error fetching documents: %v", err)
	}
	if err = cursor.All(ctx, results); err != nil {
		return fmt.Errorf("error decoding documents: %v", err)
	}
	return nil
}

// // Returns the total number of documents matching the filter
func (c *Collection[model]) Count(filter bson.M, ctx context.Context) (int64, error) {
	return c.collection.CountDocuments(ctx, filter)
}
func (c *Collection[model]) FindAndCount(filter bson.M, results interface{}, opts *options.FindOptionsBuilder, ctx context.Context) (int64, error) {
	err := c.Find(filter, results, opts, ctx)
	if err != nil {
		return 0, err
	}
	count, err2 := c.Count(filter, ctx)
	return count, err2
}
func (c *Collection[model]) All(results interface{}, opts *options.FindOptionsBuilder, ctx context.Context) error {
	filter := bson.M{}
	return c.Find(filter, results, opts, ctx)
}

// // Returns documents and the total number of documents matching the filter
// func (c *Collection[model]) FindAndCount(filter bson.M, opts *options.FindOptionsBuilder) ([]*model, int64, error) {
// 	results, err := c.Find(filter, opts)
// 	if err != nil {
// 		return results, 0, err
// 	}
// 	count, err := c.Count(filter)
// 	return results, count, err
// }

// // Returns all documents in the collection
// func (c *Collection[model]) All(opts *options.FindOptionsBuilder) ([]*model, error) {
// 	filter := bson.M{}
// 	return c.Find(filter, opts)
// }

// // Deletes all documents in the collection
// func (c *Collection[model]) Clear() error {
// 	filter := bson.M{}
// 	res, err := c.collection.DeleteMany(ctx, filter, nil)
// 	fmt.Println("Clear result: ", res)
// 	return err
// }

// // Updates documents matching the filter
// func (c *Collection[model]) Update(filter bson.M, update bson.M, opts *options.UpdateManyOptionsBuilder) error {
// 	res, err := c.collection.UpdateMany(ctx, filter, update, opts)
// 	fmt.Println("Update result: ", res)
// 	return err
// }

// // Deletes documents matching the filter
// func (c *Collection[model]) Delete(filter bson.M, opts *options.DeleteManyOptionsBuilder) error {
// 	res, err := c.collection.DeleteMany(ctx, filter, opts)
// 	fmt.Println("Delete result: ", res)
// 	return err
// }
