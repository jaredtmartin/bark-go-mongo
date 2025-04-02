package bark

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrObjNotFound = errors.New("object not found")

type Model interface {
	// Returns the unique identifier for the model
	GetId() string
	// Sets the unique identifier for the model
	SetId(id string)
	// Returns the collection for the model
	Collection(ctx context.Context) (*mongo.Collection, error)
	GetCollectionName() string
	SetCollectionName(name string)
	// Returns all the documents matching the filter
	Find(filter bson.M, opts *options.FindOptionsBuilder, ctx context.Context) ([]any, error)
	// Returns a single document matching the filter
	FindOne(filter bson.M, ctx context.Context) (*any, error)
	// Returns the document from the db matching the id
	Get(id string, ctx context.Context) (*any, error)
	// Returns the obj from the db with the id matching the models id
	Load(ctx context.Context) any
	// Returns the total number of documents matching the filter
	Count(filter bson.M, ctx context.Context)
	// Returns the total number of documents matching the filter and the documents
	FindAndCount(filter bson.M, opts *options.FindOptionsBuilder, ctx context.Context)
	// Returns all documents in the collection
	All(opts *options.FindOptionsBuilder, ctx context.Context) ([]any, error)
	// Saves the object to the database
	Save(ctx context.Context) error
	// Deletes the object from the database
	Delete(ctx context.Context) error
	// Deletes all documents matching the filter
	DeleteMany(filter bson.M, ctx context.Context) error
	// A base method to be used by models to saves the model to the database
	SaveModel(obj any, ctx context.Context) error
}

// Base model to be embedded in all models
type DefaultModel[T any] struct {
	collection     *mongo.Collection
	CollectionName string    `json:"-" bson:"-"`
	ID             string    `json:"_id" bson:"_id,omitempty"`
	Id             string    `json:"Id" bson:"Id,omitempty"`
	CreatedOn      time.Time `json:"CreatedOn" bson:"CreatedOn,omitempty"`
	UpdatedOn      time.Time `json:"UpdatedOn" bson:"UpdatedOn,omitempty"`
	Version        int       `json:"Version" bson:"Version,omitempty"`
}

// Returns the collection name
func (m *DefaultModel[T]) GetCollectionName() string {
	return m.CollectionName
}

// Sets the collection name
func (m *DefaultModel[T]) SetCollectionName(name string) {
	m.CollectionName = name
}

// Returns the unique identifier for the model
func (m *DefaultModel[T]) GetId() string {
	return m.Id
}

// Sets the unique identifier for the model
func (m *DefaultModel[T]) SetId(id string) {
	m.Id = id
	m.ID = id
}

// Returns the collection for the model
func (m *DefaultModel[T]) Collection(ctx context.Context) (*mongo.Collection, error) {
	// If you already have a collection cached, return it
	if m.collection != nil {
		return m.collection, nil
	}
	// If there's no collectionName, return an error
	if m.CollectionName == "" {
		return nil, fmt.Errorf("CollectionName not set")
	}
	// Get the database connection
	database, err := Db(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting collection %s for: %v", m.CollectionName, err)
	}
	// Get the collection
	m.collection = database.Collection(m.CollectionName)
	return m.collection, nil
}

// Returns all documents matching the filter
func (m *DefaultModel[T]) Find(filter bson.M, opts *options.FindOptionsBuilder, ctx context.Context) ([]T, error) {
	var results []T
	collection, err := m.Collection(ctx)
	if err != nil {
		return results, fmt.Errorf("failed to get collection to save model to: %v", err)
	}
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return results, fmt.Errorf("error fetching documents: %v", err)
	}
	if err = cursor.All(ctx, &results); err != nil {
		return results, fmt.Errorf("error decoding documents: %v", err)
	}
	return results, nil
}

// Returns a single document from the db matching the filter
func (m *DefaultModel[T]) FindOne(filter bson.M, ctx context.Context) (*T, error) {
	collection, err := m.Collection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection to save model to: %v", err)
	}
	var obj *T
	// create a new instance of the model with the collection name set
	// obj := DefaultModel[T]{}
	// obj.SetCollectionName(collection.Name())
	err = collection.FindOne(ctx, filter).Decode(obj)
	if err != nil {
		return nil, fmt.Errorf("error fetching documents: %v", err)
	}
	return obj, nil
}

// Returns the document from the db matching the id
func (m *DefaultModel[T]) Get(id string, ctx context.Context) (*T, error) {
	filter := bson.M{"Id": id}
	return m.FindOne(filter, ctx)
}

// Returns the document from the db by the id
func (m *DefaultModel[T]) Load(ctx context.Context) (*T, error) {
	obj, err := m.Get(m.Id, ctx)
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil, ErrObjNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching documents: %v", err)
	}
	return obj, nil
	// I can't figure out how to cast the obj to DefaultModel[T]
	// but I'll leave this here just in case I want to try again in the future
	// // m = t.DefaultModel[T]
	// // if castedObj, ok := any(t.Def).(DefaultModel[T]); ok {
	// // 	*m = castedObj
	// // } else {
	// // 	return fmt.Errorf("type assertion failed: cannot cast obj to DefaultModel[T]")
	// // }
	// return nil
}

// Returns the total number of documents matching the filter
func (m *DefaultModel[T]) Count(filter bson.M, ctx context.Context) (int64, error) {
	collection, err := m.Collection(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get collection to count documents: %v", err)
	}
	return collection.CountDocuments(ctx, filter)
}

// Returns the total number of documents matching the filter and the documents
func (m *DefaultModel[T]) FindAndCount(filter bson.M, opts *options.FindOptionsBuilder, ctx context.Context) (int64, error) {
	var results []*DefaultModel[T]
	collection, err := m.Collection(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get collection to fetch and count documents: %v", err)
	}
	err = Find(collection, filter, results, opts, ctx)
	if err != nil {
		return 0, err
	}
	count, err := Count(collection, filter, ctx)
	return count, err
}

// Returns all documents in the collection
func (m *DefaultModel[T]) All(opts *options.FindOptionsBuilder, ctx context.Context) ([]*DefaultModel[T], error) {
	var results []*DefaultModel[T]
	filter := bson.M{}
	collection, err := m.Collection(ctx)
	if err != nil {
		return results, fmt.Errorf("failed to get collection to fetch all documents: %v", err)
	}
	err = Find(collection, filter, results, opts, ctx)
	return results, err
}

// A base method to be used by models to saves the model to the database
func (m *DefaultModel[T]) SaveModel(obj any, ctx context.Context) error {
	collection, err := m.Collection(ctx)
	if err != nil {
		return fmt.Errorf("failed to get collection to save model to: %v", err)
	}
	id := m.GetId()
	if id == "" {
		m.SetId(Uuid())
	} else {
		m.SetId(id)
	}
	filter := bson.M{"Id": id}
	update := bson.M{
		"$set": obj,
		"$inc": bson.M{"Version": 1},
		"$setOnInsert": bson.M{
			"CreatedOn": Now(ctx),
		},
	}
	opts := options.UpdateOne().SetUpsert(true)
	res, err := collection.UpdateOne(ctx, filter, update, opts)
	fmt.Println("Saved: Matched:", res.MatchedCount, " Modified: ", res.ModifiedCount, " Upserted: ", res.UpsertedCount, " UpsertedID: ", res.UpsertedID)
	if err != nil {
		return fmt.Errorf("error saving model: %v", err)
	}
	return nil
}

// Deletes the object from the database
func (m *DefaultModel[T]) Delete(ctx context.Context) error {
	collection, err := m.Collection(ctx)
	if err != nil {
		return fmt.Errorf("failed to get collection to delete model from: %v", err)
	}
	id := m.GetId()
	if id == "" {
		return fmt.Errorf("cannot delete model with no id")
	}
	filter := bson.M{"Id": id}
	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("error deleting model: %v", err)
	} else {
		fmt.Println("Delete result: ", res)
	}
	return nil
}

// Deletes all documents matching the filter
func (m *DefaultModel[T]) DeleteMany(filter bson.M, ctx context.Context) error {
	collection, err := m.Collection(ctx)
	if err != nil {
		return fmt.Errorf("error getting collection to clear: %v", err)
	}
	_, err = collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("error deleting documents: %v", err)
	}
	return nil
}
