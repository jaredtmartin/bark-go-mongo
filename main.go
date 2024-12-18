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

var ErrMissingId = errors.New("missing id")

type Model interface {
	GetId() string
	SetId(id string)
	ClearVersion()
}

type DefaultModel struct {
	Id        string    `json:"Id" bson:"Id,omitempty"`
	CreatedOn time.Time `json:"CreatedOn" bson:"CreatedOn"`
	UpdatedOn time.Time `json:"UpdatedOn" bson:"UpdatedOn"`
	Version   int       `json:"Version" bson:"Version"`
}

func (m *DefaultModel) GetId() string {
	return m.Id
}
func (m *DefaultModel) SetId(id string) {
	m.Id = id
}
func (m *DefaultModel) ClearVersion() {
	m.Version = 0
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
	model.ClearVersion()
	ctx := context.Background()
	filter := bson.M{"Id": model.GetId()}
	res, err := collection.UpdateOne(ctx, filter, bson.M{
		"$set": model,
		"$inc": bson.M{"Version": 1},
		"$setOnInsert": bson.M{
			"CreatedOn": time.Now(),
			// "Version":   1,
		},
	}, opts)
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
