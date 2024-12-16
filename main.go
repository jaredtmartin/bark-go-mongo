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
	GetCreatedOn() time.Time
	SetCreatedOn(createdOn time.Time)
	GetVersion() int
	IncrementVersion()
}

type DefaultModel struct {
	Id        string    `json:"_id" bson:"_id,omitempty"`
	CreatedOn time.Time `json:"createdOn" bson:"createdOn"`
	UpdatedOn time.Time `json:"updatedOn" bson:"updatedOn"`
	Version   int       `json:"version" bson:"version"`
	// CollectionName string
}

func (m *DefaultModel) GetId() string {
	return m.Id
}
func (m *DefaultModel) SetId(id string) {
	m.Id = id
}

func (m *DefaultModel) GetCreatedOn() time.Time {
	return m.CreatedOn
}
func (m *DefaultModel) SetCreatedOn(createdOn time.Time) {
	m.CreatedOn = createdOn
}
func (m *DefaultModel) IncrementVersion() {
	m.Version += 1
}
func (m *DefaultModel) GetVersion() int {
	return m.Version
}

// func (m *DefaultModel) Delete(collection mongo.Collection) {
// I removed this becuase it seems better to add this function to each model
// 	ctx := context.Background()
// 	filter := bson.M{"_id": m.Id}
// 	collection.DeleteOne(ctx, filter)
// }

// I cant seem to make this work because the dEfaultModel doesnt have all the fields
// func (m *DefaultModel) Save(c *fiber.Ctx) error {
// 	return Save(c, m.CollectionName, m)
// }

func Get(collection *mongo.Collection, obj Model) error {
	id := obj.GetId()
	if id == "" {
		return ErrMissingId
	}
	filter := bson.M{"_id": obj.GetId()}
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
func Save(collection *mongo.Collection, model Model, opts *options.UpdateOptionsBuilder) error {
	if model.GetId() == "" {
		model.SetId(Uuid())
	}
	if model.GetCreatedOn().IsZero() {
		return Insert(model, collection, nil)
	}
	return Update(model, collection, opts)
}
func Insert(model Model, collection *mongo.Collection, opts *options.InsertOneOptionsBuilder) error {
	fmt.Printf("Inserting %v\n", model)
	model.SetCreatedOn(time.Now())
	model.IncrementVersion()
	ctx := context.Background()

	res, err := collection.InsertOne(ctx, model, opts)

	if err != nil {
		fmt.Println("mongo insert err", err)
		return err
	}
	fmt.Println("res", res)
	// model.SetId(res.InsertedID.(string))

	return nil
}
func Update(model Model, collection *mongo.Collection, opts *options.UpdateOptionsBuilder) error {
	fmt.Printf("Saving %v\n", model)
	ctx := context.Background()
	filter := bson.M{"_id": model.GetId()}
	model.IncrementVersion()
	res, err := collection.UpdateOne(ctx, filter, bson.M{"$set": model}, opts)
	fmt.Println("res", res)
	// fmt.Println("err", err)

	if err != nil {
		return err
	}
	return nil
}

func Delete(model Model, collection *mongo.Collection, opts *options.DeleteOptionsBuilder) error {
	fmt.Printf("Deleting %v\n", model)
	ctx := context.Background()
	filter := bson.M{"_id": model.GetId()}
	res, err := collection.DeleteOne(ctx, filter, opts)
	fmt.Println("delete result: ", res)
	return err
}
