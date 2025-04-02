package bark_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jaredtmartin/bark-go-mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Sets environment variables and prepares a context for testing
func setupTest(name string, now string, t *testing.T) context.Context {
	ctx := context.WithValue(context.Background(), bark.DbNameKey, "test-"+name)
	ctx = context.WithValue(ctx, bark.NowKey, now)
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test")
	t.Setenv("ENV", "test")
	return ctx
}

// A Sample model for testing
type sampleModel struct {
	bark.DefaultModel[sampleModel] `bson:",inline"`
	Name                           string `json:"Name" bson:"Name,omitempty"`
}

const sampleModelCollectionName = "samples"

// Create a new sample model
func NewSampleModel(Name string) *sampleModel {
	return &sampleModel{
		DefaultModel: bark.DefaultModel[sampleModel]{
			CollectionName: sampleModelCollectionName,
		},
		Name: Name,
	}
}

// Save the model to the database
func (m *sampleModel) Save(ctx context.Context) error {
	return m.SaveModel(m, ctx)
}

// // Sets the Id on the obj and returns the obj builder style.
// // This is not necessary, but it's convienient
// func (m *sampleModel) Id(id string) *sampleModel {
// 	m.SetId(id)
// 	return m
// }

// A simple struct to recieve a name and id of objects to setup in a fixture
type Obj struct {
	Name string
	Id   string
}

func SetupFixture(fix []*Obj, ctx context.Context) (*sampleModel, error) {
	err := NewSampleModel("").DeleteMany(bson.M{}, ctx)
	if err != nil {
		return nil, fmt.Errorf("error setting up fixture: %v", err)
	}
	for _, obj := range fix {
		model := NewSampleModel(obj.Name)
		if obj.Id != "" {
			model.SetId(obj.Id)
		}
		err := model.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("error setting up fixture %s: %v", obj.Name, err)
		}
	}
	return &sampleModel{
		DefaultModel: bark.DefaultModel[sampleModel]{
			CollectionName: sampleModelCollectionName,
		},
	}, nil
}
