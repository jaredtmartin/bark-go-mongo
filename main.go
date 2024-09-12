package bark

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
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
	GetCollectionName() string
	SetCollectionName(collection string)
	GetVersion() int
	IncrementVersion()
}

type DefaultModel struct {
	Id             string    `json:"_id" bson:"_id,omitempty"`
	CreatedOn      time.Time `json:"createdOn" bson:"createdOn"`
	UpdatedOn      time.Time `json:"updatedOn" bson:"updatedOn"`
	Version        int       `json:"version" bson:"version"`
	CollectionName string
}

func (m *DefaultModel) GetCollectionName() string {
	return m.CollectionName
}
func (m *DefaultModel) SetCollectionName(collection_name string) {
	m.CollectionName = collection_name
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

// I cant seem to make this work because the dEfaultModel doesnt have all the fields
//
//	func (m *DefaultModel) Save(c *fiber.Ctx) error {
//		return Save(c, m.CollectionName, m)
//	}

func Get(c *fiber.Ctx, collection_name string, obj Model) error {
	id := obj.GetId()
	if id == "" {
		return ErrMissingId
	}
	filter := bson.M{"_id": obj.GetId()}
	// fmt.Println("filter", filter)

	return FindOne(c, collection_name, filter, obj)
}
func FindOne(c *fiber.Ctx, collection_name string, filter bson.M, obj Model) error {
	ctx := context.Background()
	db := c.Locals("db").(*mongo.Database)
	collection := db.Collection(collection_name)
	err := collection.FindOne(ctx, filter).Decode(obj)
	if err != nil {
		return err
	}
	obj.SetCollectionName(collection_name)
	return nil
}
func Find(c *fiber.Ctx, collection_name string, filter bson.M, results interface{}, opts *options.FindOptionsBuilder) error {
	ctx := context.Background()
	db := c.Locals("db").(*mongo.Database)
	collection := db.Collection(collection_name)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	if err = cursor.All(ctx, results); err != nil {
		return err
	}
	// I don't know why this doesnt work
	// models, ok := results.([]Model)
	// if !ok {
	// 	log.Println("results is not of type []*Model")
	// } else {
	// 	for _, obj := range models {
	// 		obj.SetCollectionName(collection_name)
	// 	}
	// }
	return nil
}
func FindAndCount(c *fiber.Ctx, collection_name string, filter bson.M, results interface{}, opts *options.FindOptionsBuilder) (int64, error) {
	err := Find(c, collection_name, filter, results, opts)
	if err != nil {
		return 0, err
	}
	count, err2 := Count(c, collection_name, filter)
	return count, err2
}
func Count(c *fiber.Ctx, collection_name string, filter bson.M) (int64, error) {
	db := c.Locals("db").(*mongo.Database)
	collection := db.Collection(collection_name)
	return collection.CountDocuments(context.Background(), filter)
}
func All(c *fiber.Ctx, collection_name string, results interface{}, opts *options.FindOptionsBuilder) error {
	filter := bson.M{}
	return Find(c, collection_name, filter, results, opts)
}
func Save(c *fiber.Ctx, model Model, opts *options.UpdateOptionsBuilder) error {
	if model.GetId() == "" {
		model.SetId(Uuid(c))
	}
	db := c.Locals("db").(*mongo.Database)
	collection := db.Collection(model.GetCollectionName())
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
func Delete(ctx *fiber.Ctx, collection_name string, id string) error {
	db := ctx.Locals("db").(*mongo.Database)
	collection := db.Collection(collection_name)
	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}
func DeleteMany(ctx *fiber.Ctx, collection_name string, filter bson.M) error {
	db := ctx.Locals("db").(*mongo.Database)
	collection := db.Collection(collection_name)
	_, err := collection.DeleteMany(context.Background(), filter)
	return err
}
