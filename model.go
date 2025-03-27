package bark

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Model interface {
	// Returns the unique identifier for the model
	GetId() string
	// Sets the unique identifier for the model
	SetId(id string)

	// Returns the collection for the model
	Collection(ctx context.Context) (*mongo.Collection, error)
	// Saves the object to the database
	Save(ctx context.Context) error
	// Deletes the object from the database
	Delete(ctx context.Context) error
}

// Base model to be embedded in all models
type DefaultModel struct {
	collection *mongo.Collection
	ID         string    `json:"_id" bson:"_id,omitempty"`
	Id         string    `json:"Id" bson:"Id,omitempty"`
	CreatedOn  time.Time `json:"CreatedOn" bson:"CreatedOn,omitempty"`
	UpdatedOn  time.Time `json:"UpdatedOn" bson:"UpdatedOn,omitempty"`
	Version    int       `json:"Version" bson:"Version,omitempty"`
}

// Returns the unique identifier for the model
func (m *DefaultModel) GetId() string {
	return m.Id
}

// Sets the unique identifier for the model
func (m *DefaultModel) SetId(id string) {
	m.Id = id
	m.ID = id
}

// Base method to get the collection for any model.
// Each model should implement a more specfic Collection method which in turn calls this method
func (m *DefaultModel) GetMongoCollection(name string, ctx context.Context) (*mongo.Collection, error) {
	if m.collection == nil {
		database, err := Db(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting collection %s for model: %v", name, err)
		}
		m.collection = database.Collection(name)
	}
	return m.collection, nil
}

func Save(model Model, ctx context.Context) error {
	// We don't ask for the collection, because we need the model and ctx here anyway,
	// so we can get the collection ourselves.
	collection, err := model.Collection(ctx)
	if err != nil {
		return fmt.Errorf("failed to get collection to save model to: %v", err)
	}
	if model.GetId() == "" {
		model.SetId(Uuid())
	}
	filter := bson.M{"Id": model.GetId()}
	update := bson.M{
		"$set": model,
		"$inc": bson.M{"Version": 1},
		"$setOnInsert": bson.M{
			"CreatedOn": Now(),
		},
	}
	opts := options.UpdateOne().SetUpsert(true)
	res, err := collection.UpdateOne(ctx, filter, update, opts)
	fmt.Println("Saved: Matched:", res.MatchedCount, " Modified: ", res.ModifiedCount, " Upserted: ", res.UpsertedCount, " UpsertedID: ", res.UpsertedID)

	if err != nil {
		return err
	}
	return nil
}
func Delete(model Model, ctx context.Context) error {
	// We don't ask for the collection, because we need the model and ctx here anyway,
	// so we can get the collection ourselves.
	fmt.Printf("Deleting %v\n", model)
	collection, err := model.Collection(ctx)
	if err != nil {
		return fmt.Errorf("failed to get collection to save model to: %v", err)
	}
	filter := bson.M{"Id": model.GetId()}
	res, err := collection.DeleteOne(ctx, filter, nil)
	fmt.Println("Delete result: ", res)
	return err
}
