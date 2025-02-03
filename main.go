package bark

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrMissingId = errors.New("missing id")

type Model interface {
	GetId() string
	SetId(id string)
	ClearBaseFields()
}

type DefaultModel struct {
	ID        string    `json:"_id" bson:"_id,omitempty"`
	Id        string    `json:"Id" bson:"Id,omitempty"`
	CreatedOn time.Time `json:"CreatedOn" bson:"CreatedOn,omitempty"`
	UpdatedOn time.Time `json:"UpdatedOn" bson:"UpdatedOn,omitempty"`
	Version   int       `json:"Version" bson:"Version,omitempty"`
}

func (m *DefaultModel) GetId() string {
	return m.Id
}
func (m *DefaultModel) SetId(id string) {
	m.Id = id
}
func (m *DefaultModel) ClearBaseFields() {
	m.ID = m.Id
	m.Version = 0
	m.CreatedOn = time.Time{}
	m.UpdatedOn = time.Now()
}

// I cant seem to make this work because the dEfaultModel doesnt have all the fields
// func (m *DefaultModel) Save(c *fiber.Ctx) error {
// 	return Save(c, m.CollectionName, m)
// }

func Get(collection *mongo.Collection, obj Model) error {
	id := obj.GetId()
	if id == "" {
		return ErrMissingId
	}
	filter := bson.M{"Id": obj.GetId()}
	// fmt.Println("filter", filter)

	return FindOne(collection, filter, obj)
}
func FindOne(collection *mongo.Collection, filter bson.M, obj Model) error {
	ctx := context.Background()
	err := collection.FindOne(ctx, filter).Decode(obj)
	if err != nil {
		return err
	}
	return nil
}
func Find(collection *mongo.Collection, filter bson.M, results interface{}, opts *options.FindOptionsBuilder) error {
	ctx := context.Background()

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	if err = cursor.All(ctx, results); err != nil {
		return err
	}
	return nil
}
func FindAndCount(collection *mongo.Collection, filter bson.M, results interface{}, opts *options.FindOptionsBuilder) (int64, error) {
	err := Find(collection, filter, results, opts)
	if err != nil {
		return 0, err
	}
	count, err2 := Count(collection, filter)
	return count, err2
}
func Count(collection *mongo.Collection, filter bson.M) (int64, error) {
	return collection.CountDocuments(context.Background(), filter)
}
func All(collection *mongo.Collection, results interface{}, opts *options.FindOptionsBuilder) error {
	filter := bson.M{}
	return Find(collection, filter, results, opts)
}

func Save(model Model, collection *mongo.Collection, opts *options.UpdateOptionsBuilder) error {
	fmt.Printf("Saving %v\n", model)
	if model.GetId() == "" {
		model.SetId(Uuid())
	}
	model.ClearBaseFields()
	ctx := context.Background()
	filter := bson.M{"Id": model.GetId()}
	fmt.Println("model", model)
	update := bson.M{
		"$set": model,
		"$inc": bson.M{"Version": 1},
		"$setOnInsert": bson.M{
			"CreatedOn": time.Now(),
		},
	}
	log.Println("update", update)
	res, err := collection.UpdateOne(ctx, filter, update, opts)
	fmt.Println(" ", res.MatchedCount, res.ModifiedCount, res.UpsertedCount, res.UpsertedID)
	fmt.Println("Save result: ", res)
	// fmt.Println("err", err)

	if err != nil {
		return err
	}
	return nil
}

func Delete(model Model, collection *mongo.Collection, opts *options.DeleteOptionsBuilder) error {
	fmt.Printf("Deleting %v\n", model)
	ctx := context.Background()
	filter := bson.M{"Id": model.GetId()}
	res, err := collection.DeleteOne(ctx, filter, opts)
	fmt.Println("Delete result: ", res)
	return err
}
