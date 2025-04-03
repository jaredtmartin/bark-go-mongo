package bark_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jaredtmartin/bark-go-mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
type Dog struct {
	bark.Model `bson:",inline"`
	Name       string `json:"Name" bson:"Name,omitempty"`
	Age        int    `json:"Age" bson:"Age,omitempty"`
}

const DogCollectionName = "dogs"

// Create a new sample model
func NewDog(Name string) *Dog {
	return &Dog{
		Model: bark.Model{
			CollectionName: DogCollectionName,
		},
		Name: Name,
	}
}

// Save the model to the database
func (m *Dog) Save(ctx context.Context) (*bark.Result, error) {
	return m.SaveModel(m, ctx)
}
func (m *Dog) ToFixture() *Obj {
	return &Obj{
		Name: m.Name,
		Id:   m.Id,
		Age:  m.Age,
	}
}
func (m *Dog) String() string {
	return fmt.Sprintf("Dog Id: %s, Name: %s, Age: %d", m.Id, m.Name, m.Age)
}

// // Sets the Id on the obj and returns the obj builder style.
// // This is not necessary, but it's convienient
// func (m *Dog) Id(id string) *Dog {
// 	m.SetId(id)
// 	return m
// }

// A simple struct to recieve a name and id of objects to setup in a fixture
type Obj struct {
	Name string
	Age  int
	Id   string
}

func SetupFixture(fix []*Obj, ctx context.Context) (*bark.Collection[*Dog], error) {
	dogs := bark.NewCollection[*Dog](DogCollectionName)
	_, err := dogs.DeleteMany(bson.M{}, ctx)
	if err != nil {
		return nil, fmt.Errorf("error setting up fixture: %v", err)
	}
	for _, obj := range fix {
		model := NewDog(obj.Name)
		model.Id = obj.Id
		model.ID = obj.Id
		model.Age = obj.Age
		// fmt.Println("Saving model: ", model.String())
		_, err := model.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("error setting up fixture %s: %v", obj.Name, err)
		}
	}
	return dogs, nil
}
func TestCommonFind(t *testing.T) {
	ctx := setupTest("CommonFind", "2024-01-01T00:00:00Z", t)

	fixture := []*Obj{
		{Name: "Buddy", Age: 3, Id: "1"},
		{Name: "Max", Age: 5, Id: "2"},
		{Name: "Charlie", Age: 2, Id: "3"},
	}

	dogs, err := SetupFixture(fixture, ctx)
	if err != nil {
		t.Fatalf("Failed to setup fixture: %v", err)
	}
	// fmt.Println("fixture setup successfully")
	collection, err := dogs.MongoCollection(ctx)
	if err != nil {
		t.Fatalf("Failed to get collection: %v", err)
	}
	// fmt.Println("collection setup successfully")
	t.Run("Find with valid filter", func(t *testing.T) {
		var results []*Dog
		filter := bson.M{"Age": bson.M{"$gt": 2}}

		err = bark.Find(collection, filter, &results, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("Find with empty filter", func(t *testing.T) {
		var results []*Dog
		filter := bson.M{}
		err := bark.Find(collection, filter, &results, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})

	t.Run("Find with non-existing filter", func(t *testing.T) {
		var results []*Dog
		filter := bson.M{"Name": "NonExistent"}
		err := bark.Find(collection, filter, &results, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})

	t.Run("Find with invalid results parameter", func(t *testing.T) {
		var results string
		filter := bson.M{}
		err := bark.Find(collection, filter, &results, nil, ctx)
		if err == nil {
			t.Error("Expected error for invalid results parameter, got nil")
		}
	})

	t.Run("Find with complex filter", func(t *testing.T) {
		var results []*Dog
		filter := bson.M{
			"$and": []bson.M{
				{"Age": bson.M{"$gte": 2}},
				{"Age": bson.M{"$lte": 5}},
				{"Name": bson.M{"$in": []string{"Buddy", "Max"}}},
			},
		}
		err := bark.Find(collection, filter, &results, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})
}
func TestCommonCount(t *testing.T) {
	ctx := setupTest("CommonCount", "2024-01-01T00:00:00Z", t)

	fixture := []*Obj{
		{Name: "Buddy", Age: 3, Id: "1"},
		{Name: "Max", Age: 5, Id: "2"},
		{Name: "Charlie", Age: 2, Id: "3"},
	}

	dogs, err := SetupFixture(fixture, ctx)
	if err != nil {
		t.Fatalf("Failed to setup fixture: %v", err)
	}
	// fmt.Println("fixture setup successfully")
	collection, err := dogs.MongoCollection(ctx)
	if err != nil {
		t.Fatalf("Failed to get collection: %v", err)
	}
	// fmt.Println("collection setup successfully")

	t.Run("Count with valid filter", func(t *testing.T) {
		filter := bson.M{"Age": bson.M{"$gt": 2}}
		count, err := bark.Count(collection, filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count to be 2, got %d", count)
		}
	})

	t.Run("Count with empty filter", func(t *testing.T) {
		filter := bson.M{}
		count, err := bark.Count(collection, filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count to be 3, got %d", count)
		}
	})

	t.Run("Count with non-existing filter", func(t *testing.T) {
		filter := bson.M{"Name": "NonExistent"}
		count, err := bark.Count(collection, filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count to be 0, got %d", count)
		}
	})

	t.Run("Count with complex filter", func(t *testing.T) {
		filter := bson.M{
			"$and": []bson.M{
				{"Age": bson.M{"$gte": 2}},
				{"Age": bson.M{"$lte": 5}},
				{"Name": bson.M{"$in": []string{"Buddy", "Max"}}},
			},
		}
		count, err := bark.Count(collection, filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count to be 2, got %d", count)
		}
	})
}
func TestCommonFindAndCount(t *testing.T) {
	ctx := setupTest("CommonFindAndCount", "2024-01-01T00:00:00Z", t)

	fixture := []*Obj{
		{Name: "Buddy", Age: 3, Id: "1"},
		{Name: "Max", Age: 5, Id: "2"},
		{Name: "Charlie", Age: 2, Id: "3"},
		{Name: "Rocky", Age: 4, Id: "4"},
		{Name: "Luna", Age: 1, Id: "5"},
	}

	dogs, err := SetupFixture(fixture, ctx)
	if err != nil {
		t.Fatalf("Failed to setup fixture: %v", err)
	}

	collection, err := dogs.MongoCollection(ctx)
	if err != nil {
		t.Fatalf("Failed to get collection: %v", err)
	}

	t.Run("FindAndCount with age range and sort", func(t *testing.T) {
		var results []*Dog
		filter := bson.M{"Age": bson.M{"$gt": 2, "$lt": 4}}
		opts := options.Find().SetSort(bson.M{"Age": 1})
		count, err := bark.FindAndCount(collection, filter, &results, opts, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != int64(len(results)) {
			t.Errorf("Count mismatch: got count=%d but results length=%d", count, len(results))
		}
	})

	t.Run("FindAndCount with nil options", func(t *testing.T) {
		var results []*Dog
		filter := bson.M{"Age": bson.M{"$gt": 0}}
		count, err := bark.FindAndCount(collection, filter, &results, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 5 {
			t.Errorf("Expected count of 5, got %d", count)
		}
	})

	t.Run("FindAndCount with invalid results parameter", func(t *testing.T) {
		var results string
		filter := bson.M{}
		count, err := bark.FindAndCount(collection, filter, &results, nil, ctx)
		if err == nil {
			t.Error("Expected error for invalid results parameter, got nil")
		}
		if count != 0 {
			t.Errorf("Expected count of 0 for error case, got %d", count)
		}
	})

	t.Run("FindAndCount with regex filter", func(t *testing.T) {
		var results []*Dog
		filter := bson.M{"Name": bson.M{"$regex": "^[BM]"}}
		count, err := bark.FindAndCount(collection, filter, &results, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count of 2 for names starting with B or M, got %d", count)
		}
	})

	t.Run("FindAndCount with limit option", func(t *testing.T) {
		var results []*Dog
		filter := bson.M{}
		opts := options.Find().SetLimit(3)
		count, err := bark.FindAndCount(collection, filter, &results, opts, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 5 {
			t.Errorf("Expected total count of 5 despite limit, got %d", count)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results due to limit, got %d", len(results))
		}
	})
}
func TestCommonAll(t *testing.T) {
	ctx := setupTest("CommonAll", "2024-01-01T00:00:00Z", t)

	fixture := []*Obj{
		{Name: "Buddy", Age: 3, Id: "1"},
		{Name: "Max", Age: 5, Id: "2"},
		{Name: "Charlie", Age: 2, Id: "3"},
	}

	dogs, err := SetupFixture(fixture, ctx)
	if err != nil {
		t.Fatalf("Failed to setup fixture: %v", err)
	}

	collection, err := dogs.MongoCollection(ctx)
	if err != nil {
		t.Fatalf("Failed to get collection: %v", err)
	}

	t.Run("All with no options", func(t *testing.T) {
		var results []*Dog
		err := bark.All(collection, &results, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})

	t.Run("All with limit option", func(t *testing.T) {
		var results []*Dog
		opts := options.Find().SetLimit(2)
		err := bark.All(collection, &results, opts, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results due to limit, got %d", len(results))
		}
	})

	t.Run("All with invalid results parameter", func(t *testing.T) {
		var results string
		err := bark.All(collection, &results, nil, ctx)
		if err == nil {
			t.Error("Expected error for invalid results parameter, got nil")
		}
	})
}
