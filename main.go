package bark

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const tokenLength = 16

func NewUuid() string {
	b := make([]byte, tokenLength)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

var ErrMissingId = errors.New("missing id")

type Model interface {
	GetId() string
	SetId(id string)
	GetCreatedOn() time.Time
	SetCreatedOn(createdOn time.Time)
	GetUpdatedOn() time.Time
	SetUpdatedOn(updatedOn time.Time)
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
func (m *DefaultModel) GetUpdatedOn() time.Time {
	return m.UpdatedOn
}
func (m *DefaultModel) SetUpdatedOn(updatedOn time.Time) {
	m.UpdatedOn = updatedOn
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

func Get(db *mongo.Database, model Model) error {
	id := model.GetId()
	if id == "" {
		return ErrMissingId
	}
	collection, err := getCollectionForModel(db, model)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": model.GetId()}
	// fmt.Println("filter", filter)
	return FindOne(collection, filter, model)
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
	return Find(collection, bson.M{}, results, opts)
}

func getCollectionForModel(db *mongo.Database, model Model) (*mongo.Collection, error) {
	name := model.GetCollectionName()
	if name == "" {
		return nil, errors.New("collection name is required to save object")
	}
	return db.Collection(model.GetCollectionName()), nil
}
func Save(db *mongo.Database, model Model, opts *options.UpdateOptionsBuilder) error {
	if model.GetId() == "" {
		model.SetId(NewUuid())
	}
	if model.GetCreatedOn().IsZero() {
		return Insert(db, model, nil)
	}
	return Update(db, model, opts)
}

func Insert(db *mongo.Database, model Model, opts *options.InsertOneOptionsBuilder) error {
	// fmt.Printf("Inserting %v\n", model)
	collection, err := getCollectionForModel(db, model)
	if err != nil {
		return err
	}
	model.SetCreatedOn(time.Now())
	model.IncrementVersion()
	ctx := context.Background()

	_, err = collection.InsertOne(ctx, model, opts)

	if err != nil {
		fmt.Println("mongo insert err", err)
		return err
	}
	// fmt.Println("res", res)
	// model.SetId(res.InsertedID.(string))

	return nil
}
func Update(db *mongo.Database, model Model, opts *options.UpdateOptionsBuilder) error {
	// fmt.Printf("Updating %v\n", model)
	collection, err := getCollectionForModel(db, model)
	if err != nil {
		return err
	}
	ctx := context.Background()
	filter := bson.M{"_id": model.GetId()}
	model.SetUpdatedOn(time.Now())
	model.IncrementVersion()
	_, err = collection.UpdateOne(ctx, filter, bson.M{"$set": model}, opts)
	// fmt.Println("res", res)
	// fmt.Println("err", err)

	if err != nil {
		return err
	}
	return nil
}
func Delete(db *mongo.Database, model Model) error {
	collection, err := getCollectionForModel(db, model)
	if err != nil {
		return err
	}
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": model.GetId()})
	return err
}
func DeleteMany(db *mongo.Database, collection_name string, filter bson.M) error {
	if collection_name == "" {
		return errors.New("collection name is required to delete many")
	}
	collection := db.Collection(collection_name)
	_, err := collection.DeleteMany(context.Background(), filter)
	return err
}
func Debug(text string, m Model) {
	text += " "
	bsonData, err := bson.Marshal(m)
	if err != nil {
		fmt.Println("Error marshaling struct:", err)
		return
	}
	var doc bson.M
	if err := bson.Unmarshal(bsonData, &doc); err != nil {
		fmt.Println("Error unmarshaling BSON to bson.M:", err)
		return
	}
	for key, value := range doc {
		text += fmt.Sprintf("%s: %v\n", key, value)
	}
	log.Println(text)
}
