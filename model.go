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

// Base model to be embedded in all models
type Model struct {
	collection     *Collection[*Model] `json:"-" bson:"-"`
	CollectionName string              `json:"-" bson:"-"`
	ID             string              `json:"_id" bson:"_id,omitempty"`
	Id             string              `json:"Id" bson:"Id,omitempty"`
	CreatedOn      time.Time           `json:"CreatedOn" bson:"CreatedOn,omitempty"`
	UpdatedOn      time.Time           `json:"UpdatedOn" bson:"UpdatedOn,omitempty"`
	Version        int                 `json:"Version" bson:"Version,omitempty"`
}

// Creates a new model
// The collection name is required
// The id is optional, if not provided, a new ID will be generated
func NewModel(collectionName string, id ...string) *Model {
	var Id string
	if len(id) > 0 {
		Id = id[0]
	}
	return &Model{
		CollectionName: collectionName,
		ID:             Id,
		Id:             Id,
	}
}

// Simple struct to report the result of an operation
type Result struct {
	Matched  int64
	Modified int64
	Inserted int64
	Deleted  int64
}

// Returns readable report of the result
func (r Result) String() string {
	return fmt.Sprintf("Matched: %d, Modified: %d, Inserted: %d, Deleted: %d", r.Matched, r.Modified, r.Inserted, r.Deleted)
}

// Returns an empty result
func EmptyResult() *Result {
	return &Result{}
}

// Returns a result from the mongo update operation
func ResultFromUpdate(result *mongo.UpdateResult) *Result {
	return &Result{Matched: result.MatchedCount, Modified: result.ModifiedCount, Inserted: result.UpsertedCount}
}

// Returns a result from the mongo delete operation
func ResultFromDelete(result *mongo.DeleteResult) *Result {
	return &Result{Deleted: result.DeletedCount}
}

// Returns the collection for the model
func (m *Model) Collection() *Collection[*Model] {
	if m.collection == nil {
		m.collection = NewCollection[*Model](m.CollectionName)
	}
	return m.collection
}

// Sets the collection name for the model
func (m *Model) SetCollectionName(name string) {
	m.CollectionName = name
}

// Gets the collection name from the model
func (m *Model) GetCollectionName() string {
	return m.CollectionName
}

// A base method to be used by models to saves the model to the database
func (m *Model) SaveModel(obj any, ctx context.Context) (*Result, error) {
	collection, err := m.Collection().MongoCollection(ctx)
	if err != nil {
		return EmptyResult(), fmt.Errorf("failed to get collection to save model to: %v", err)
	}
	if m.Id == "" {
		m.Id = Uuid()
	}
	m.ID = m.Id
	filter := bson.M{"Id": m.Id}
	// Here we convert the object to a bson map so we can make adjustments
	// We need to remove the _id field so it doesnt clash with the setOnInsert
	// We also make sure the Id field is set with the UUID we generated
	bsonMap := bson.M{}
	bsonBytes, err := bson.Marshal(obj)
	if err != nil {
		return EmptyResult(), fmt.Errorf("failed to marshal model to bson: %v", err)
	}
	bson.Unmarshal(bsonBytes, &bsonMap)
	delete(bsonMap, "_id")
	bsonMap["Id"] = m.Id
	// fmt.Println("bsonMap: ", bsonMap)
	update := bson.M{
		"$set": bsonMap,
		"$inc": bson.M{"Version": 1},
		"$setOnInsert": bson.M{
			"CreatedOn": Now(ctx),
			"_id":       m.Id,
		},
	}
	// fmt.Println("Update: ", update)
	opts := options.UpdateOne().SetUpsert(true)
	res, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return EmptyResult(), fmt.Errorf("error saving model: %v", err)
	}
	// fmt.Println("Saved: Matched:", res.MatchedCount, " Modified: ", res.ModifiedCount, " Upserted: ", res.UpsertedCount, " UpsertedID: ", res.UpsertedID)
	return ResultFromUpdate(res), nil
}

// Deletes the object from the database
func (m *Model) Delete(ctx context.Context) (*Result, error) {
	collection, err := m.Collection().MongoCollection(ctx)
	if err != nil {
		return EmptyResult(), fmt.Errorf("failed to get collection to delete model from: %v", err)
	}
	if m.Id == "" {
		return EmptyResult(), fmt.Errorf("cannot delete model with no id")
	}
	filter := bson.M{"Id": m.Id}
	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return EmptyResult(), fmt.Errorf("error deleting model: %v", err)
	}
	return ResultFromDelete(res), nil
}
