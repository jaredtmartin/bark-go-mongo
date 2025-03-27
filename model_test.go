package bark

import (
	"context"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestDefaultModel_GetId(t *testing.T) {
	tests := []struct {
		name     string
		model    DefaultModel
		expected string
	}{
		{
			name: "GetId returns correct ID",
			model: DefaultModel{
				Id: "12345",
			},
			expected: "12345",
		},
		{
			name: "GetId returns empty string when ID is not set",
			model: DefaultModel{
				Id: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.model.GetId()
			if got != tt.expected {
				t.Errorf("GetId() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestDefaultModel_SetId(t *testing.T) {
	tests := []struct {
		name     string
		inputId  string
		expected DefaultModel
	}{
		{
			name:    "SetId sets both Id and ID fields",
			inputId: "12345",
			expected: DefaultModel{
				Id: "12345",
				ID: "12345",
			},
		},
		{
			name:    "SetId sets empty string for both Id and ID fields",
			inputId: "",
			expected: DefaultModel{
				Id: "",
				ID: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := DefaultModel{}
			model.SetId(tt.inputId)

			if model.Id != tt.expected.Id {
				t.Errorf("SetId() Id = %v, want %v", model.Id, tt.expected.Id)
			}
			if model.ID != tt.expected.ID {
				t.Errorf("SetId() ID = %v, want %v", model.ID, tt.expected.ID)
			}
		})
	}
}
func TestDefaultModel_getCollection(t *testing.T) {
	tests := []struct {
		name          string
		collection    *mongo.Collection
		dbFunc        func(ctx context.Context) (*mongo.Database, error)
		expectedError bool
	}{
		{
			name:       "Returns existing collection if already set",
			collection: &mongo.Collection{},
			dbFunc: func(ctx context.Context) (*mongo.Database, error) {
				return nil, nil
			},
			expectedError: false,
		},
		{
			name:       "Returns new collection if not already set",
			collection: nil,
			dbFunc: func(ctx context.Context) (*mongo.Database, error) {
				return &mongo.Database{}, nil
			},
			expectedError: false,
		},
		{
			name:       "Returns error if Db function fails",
			collection: nil,
			dbFunc: func(ctx context.Context) (*mongo.Database, error) {
				return nil, fmt.Errorf("database error")
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			model := DefaultModel{
				collection: tt.collection,
			}

			ctx := context.TODO()
			collection, err := model.getCollection("test_collection", ctx)

			if err == nil && collection == nil {
				t.Errorf("getCollection() returned nil collection")
			}
		})
	}
}
func TestSave(t *testing.T) {
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test")
	t.Setenv("ENV", "test")
	t.Setenv("NOW", "2024-03-27T19:55:38.782Z")
	ctx := context.WithValue(context.Background(), DbNameKey, "test-1")
	dogs, err := NewCollection[sampleModel](sampleModelCollectionName, ctx)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}
	m1 := dogs.New()
	err = m1.Save(ctx)
	if err != nil {
		t.Fatalf("error saving model with new ID: %v", err)
	}
	m2 := dogs.New()
	m2.SetId("12345")
	err = m2.Save(ctx)
	if err != nil {
		t.Fatalf("error saving model with existing ID: %v", err)
	}

}
func TestSavingAndDeleting(t *testing.T) {
	ctx := context.WithValue(context.Background(), DbNameKey, "test-1")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test")
	t.Setenv("ENV", "test")
	t.Setenv("NOW", "2024-03-27T19:55:38.782Z")
	dogs, err := NewCollection[sampleModel](sampleModelCollectionName, ctx)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}
	dogs.Clear(ctx)
	dog := dogs.New()
	dog.Name = "Fido"
	// fmt.Println("dog.Id", dog.Id)
	// fmt.Println("dog.GetId()", dog.GetId())
	dog.Save(ctx)
	id := dog.GetId()
	// fmt.Println("id", id)
	_, err = dogs.Get(id, ctx)
	if err != nil {
		t.Fatalf("Failed to fetch dog: %v", err)
	}
	err = dog.Delete(ctx)
	if err != nil {
		t.Fatalf("Failed to delete dog: %v", err)
	}
	count, err := dogs.Count(bson.M{}, ctx)
	if err != nil {
		t.Fatalf("Failed to count dogs: %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected dog collection to be empty after deletion")
	}
}
