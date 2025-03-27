package bark

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type sampleModel struct {
	DefaultModel `bson:",inline"`
	Name         string `json:"Name" bson:"Name,omitempty"`
}

const sampleModelCollectionName = "samples"

//	func (m *sampleModel) CollectionName() string {
//		return sampleModelCollectionName
//	}
func (m *sampleModel) Collection(ctx context.Context) (*mongo.Collection, error) {
	return m.GetMongoCollection(sampleModelCollectionName, ctx)
}
func (m *sampleModel) Save(ctx context.Context) error {
	return Save(m, ctx)
}
func (m *sampleModel) Delete(ctx context.Context) error {
	return Delete(m, ctx)
}
func TestSavingAndRetrieving(t *testing.T) {
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
	fectchedDog, err := dogs.Get(id, ctx)
	if err != nil {
		t.Fatalf("Failed to fetch dog: %v", err)
	}
	jsonBytes, err := json.Marshal(fectchedDog)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %s\n", err)
		return
	}
	if fectchedDog.CreatedOn.String() != "2024-03-27 19:55:38.782 +0000 UTC" {
		t.Errorf("Expected CreatedOn to be 2024-03-27 19:55:38.782 +0000 UTC, got %s", fectchedDog.CreatedOn)
	}
	expected := `{"_id":"0000000000000001","Id":"0000000000000001","CreatedOn":"2024-03-27T19:55:38.782Z","UpdatedOn":"0001-01-01T00:00:00Z","Version":1,"Name":"Fido"}`
	if string(jsonBytes) != expected {
		t.Errorf("Expected %s, \ngot %s", expected, string(jsonBytes))
	}

}
func TestNewCollection(t *testing.T) {
	ctx := context.WithValue(context.Background(), DbNameKey, "test-1")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test")
	t.Setenv("ENV", "test")

	// Test successful creation of a collection
	collectionName := "test_collection"
	collection, err := NewCollection[sampleModel](collectionName, ctx)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}
	if collection == nil {
		t.Fatalf("Expected collection to be non-nil")
	}
}
func TestCollectionGet(t *testing.T) {
	ctx := context.WithValue(context.Background(), DbNameKey, "test-1")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test")
	t.Setenv("ENV", "test")
	t.Setenv("NOW", "2024-03-27T19:55:38.782Z")

	// Create a new collection
	collection, err := NewCollection[sampleModel](sampleModelCollectionName, ctx)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Clear the collection to ensure a clean slate
	err = collection.Clear(ctx)
	if err != nil {
		t.Fatalf("Failed to clear collection: %v", err)
	}

	// Create and save a new sample model
	model := collection.New()
	model.Name = "Test Model"
	err = model.Save(ctx)
	if err != nil {
		t.Fatalf("Failed to save model: %v", err)
	}

	// Retrieve the model using its ID
	id := model.GetId()
	retrievedModel, err := collection.Get(id, ctx)
	if err != nil {
		t.Fatalf("Failed to retrieve model: %v", err)
	}

	// Verify the retrieved model matches the saved model
	if retrievedModel.Name != model.Name {
		t.Errorf("Expected Name to be %s, got %s", model.Name, retrievedModel.Name)
	}
	if retrievedModel.GetId() != model.GetId() {
		t.Errorf("Expected ID to be %s, got %s", model.GetId(), retrievedModel.GetId())
	}
}
func TestCollectionFindOne(t *testing.T) {
	ctx := context.WithValue(context.Background(), DbNameKey, "test-1")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test")
	t.Setenv("ENV", "test")
	t.Setenv("NOW", "2024-03-27T19:55:38.782Z")

	// Create a new collection
	collection, err := NewCollection[sampleModel](sampleModelCollectionName, ctx)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Clear the collection to ensure a clean slate
	err = collection.Clear(ctx)
	if err != nil {
		t.Fatalf("Failed to clear collection: %v", err)
	}

	// Create and save a new sample model
	model := collection.New()
	model.Name = "Test Model"
	err = model.Save(ctx)
	if err != nil {
		t.Fatalf("Failed to save model: %v", err)
	}

	// Use FindOne to retrieve the model using a filter
	filter := bson.M{"_id": model.GetId()}
	retrievedModel, err := collection.FindOne(filter, ctx)
	if err != nil {
		t.Fatalf("Failed to find model: %v", err)
	}

	// Verify the retrieved model matches the saved model
	if retrievedModel.Name != model.Name {
		t.Errorf("Expected Name to be %s, got %s", model.Name, retrievedModel.Name)
	}
	if retrievedModel.GetId() != model.GetId() {
		t.Errorf("Expected ID to be %s, got %s", model.GetId(), retrievedModel.GetId())
	}

	// Test case where no document matches the filter
	nonExistentFilter := bson.M{"_id": "nonexistent-id"}
	_, err = collection.FindOne(nonExistentFilter, ctx)
	if err == nil {
		t.Fatalf("Expected error when finding non-existent document, got nil")
	}
}
func TestCollectionClear(t *testing.T) {
	ctx := context.WithValue(context.Background(), DbNameKey, "test-1")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test")
	t.Setenv("ENV", "test")

	// Create a new collection
	collection, err := NewCollection[sampleModel](sampleModelCollectionName, ctx)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Add a document to the collection
	model := collection.New()
	model.Name = "Test Model"
	err = model.Save(ctx)
	if err != nil {
		t.Fatalf("Failed to save model: %v", err)
	}

	// Ensure the document exists
	retrievedModel, err := collection.Get(model.GetId(), ctx)
	if err != nil {
		t.Fatalf("Failed to retrieve model: %v", err)
	}
	if retrievedModel == nil {
		t.Fatalf("Expected model to exist before clearing")
	}

	// Clear the collection
	err = collection.Clear(ctx)
	if err != nil {
		t.Fatalf("Failed to clear collection: %v", err)
	}

	// Verify the collection is empty
	_, err = collection.Get(model.GetId(), ctx)
	if err == nil {
		t.Fatalf("Expected error when retrieving model after clearing collection, got nil")
	}
}

func TestCollectionClearNonTestDatabase(t *testing.T) {
	ctx := context.WithValue(context.Background(), DbNameKey, "prod-1")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "prod")
	t.Setenv("ENV", "prod")

	// Create a new collection
	collection, err := NewCollection[sampleModel](sampleModelCollectionName, ctx)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Attempt to clear the collection
	err = collection.Clear(ctx)
	if err == nil {
		t.Fatalf("Expected error when clearing collection in non-test database, got nil")
	}
	if err != ErrClearCanOnlyBeUsedOnDbsStartingWithTest {
		t.Fatalf("Expected error to be ErrClearCanOnlyBeUsedOnDbsStartingWithTest, got %v", err)
	}
}
func TestCollectionFind(t *testing.T) {
	ctx := context.WithValue(context.Background(), DbNameKey, "test-1")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test")
	t.Setenv("ENV", "test")

	// Create a new collection
	collection, err := NewCollection[sampleModel](sampleModelCollectionName, ctx)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Clear the collection to ensure a clean slate
	err = collection.Clear(ctx)
	if err != nil {
		t.Fatalf("Failed to clear collection: %v", err)
	}

	// Add multiple documents to the collection
	model1 := collection.New()
	model1.Name = "Model 1"
	err = model1.Save(ctx)
	if err != nil {
		t.Fatalf("Failed to save model1: %v", err)
	}

	model2 := collection.New()
	model2.Name = "Model 2"
	err = model2.Save(ctx)
	if err != nil {
		t.Fatalf("Failed to save model2: %v", err)
	}

	// Use Find to retrieve all documents
	var results []sampleModel
	filter := bson.M{}
	err = collection.Find(filter, &results, nil, ctx)
	if err != nil {
		t.Fatalf("Failed to find documents: %v", err)
	}

	// Verify the retrieved documents
	if len(results) != 2 {
		t.Fatalf("Expected 2 documents, got %d", len(results))
	}

	names := []string{results[0].Name, results[1].Name}
	if !(names[0] == "Model 1" && names[1] == "Model 2") && !(names[0] == "Model 2" && names[1] == "Model 1") {
		t.Errorf("Expected documents with names 'Model 1' and 'Model 2', got %v", names)
	}

	// Use Find with a filter to retrieve a specific document
	filter = bson.M{"Name": "Model 1"}
	results = []sampleModel{}
	err = collection.Find(filter, &results, nil, ctx)
	if err != nil {
		t.Fatalf("Failed to find documents with filter: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 document, got %d", len(results))
	}
	if results[0].Name != "Model 1" {
		t.Errorf("Expected document with Name 'Model 1', got %s", results[0].Name)
	}
}
func TestCollectionCount(t *testing.T) {
	ctx := context.WithValue(context.Background(), DbNameKey, "test-1")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test")
	t.Setenv("ENV", "test")

	// Create a new collection
	collection, err := NewCollection[sampleModel](sampleModelCollectionName, ctx)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Clear the collection to ensure a clean slate
	err = collection.Clear(ctx)
	if err != nil {
		t.Fatalf("Failed to clear collection: %v", err)
	}

	// Add multiple documents to the collection
	model1 := collection.New()
	model1.Name = "Model 1"
	err = model1.Save(ctx)
	if err != nil {
		t.Fatalf("Failed to save model1: %v", err)
	}

	model2 := collection.New()
	model2.Name = "Model 2"
	err = model2.Save(ctx)
	if err != nil {
		t.Fatalf("Failed to save model2: %v", err)
	}

	// Count all documents in the collection
	filter := bson.M{}
	count, err := collection.Count(filter, ctx)
	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count to be 2, got %d", count)
	}

	// Count documents with a specific filter
	filter = bson.M{"Name": "Model 1"}
	count, err = collection.Count(filter, ctx)
	if err != nil {
		t.Fatalf("Failed to count documents with filter: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count to be 1, got %d", count)
	}

	// Count documents with a filter that matches no documents
	filter = bson.M{"Name": "Nonexistent Model"}
	count, err = collection.Count(filter, ctx)
	if err != nil {
		t.Fatalf("Failed to count documents with non-matching filter: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count to be 0, got %d", count)
	}
}
func TestCollectionFindAndCount(t *testing.T) {
	ctx := context.WithValue(context.Background(), DbNameKey, "test-1")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test")
	t.Setenv("ENV", "test")

	// Create a new collection
	collection, err := NewCollection[sampleModel](sampleModelCollectionName, ctx)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Clear the collection to ensure a clean slate
	err = collection.Clear(ctx)
	if err != nil {
		t.Fatalf("Failed to clear collection: %v", err)
	}

	// Add multiple documents to the collection
	model1 := collection.New()
	model1.Name = "Model 1"
	err = model1.Save(ctx)
	if err != nil {
		t.Fatalf("Failed to save model1: %v", err)
	}

	model2 := collection.New()
	model2.Name = "Model 2"
	err = model2.Save(ctx)
	if err != nil {
		t.Fatalf("Failed to save model2: %v", err)
	}

	// Use FindAndCount to retrieve all documents and count them
	var results []sampleModel
	filter := bson.M{}
	count, err := collection.FindAndCount(filter, &results, nil, ctx)
	if err != nil {
		t.Fatalf("Failed to find and count documents: %v", err)
	}

	// Verify the count and retrieved documents
	if count != 2 {
		t.Errorf("Expected count to be 2, got %d", count)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 documents, got %d", len(results))
	}

	names := []string{results[0].Name, results[1].Name}
	if !(names[0] == "Model 1" && names[1] == "Model 2") && !(names[0] == "Model 2" && names[1] == "Model 1") {
		t.Errorf("Expected documents with names 'Model 1' and 'Model 2', got %v", names)
	}

	// Use FindAndCount with a filter to retrieve a specific document and count
	filter = bson.M{"Name": "Model 1"}
	results = []sampleModel{}
	count, err = collection.FindAndCount(filter, &results, nil, ctx)
	if err != nil {
		t.Fatalf("Failed to find and count documents with filter: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count to be 1, got %d", count)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 document, got %d", len(results))
	}
	if results[0].Name != "Model 1" {
		t.Errorf("Expected document with Name 'Model 1', got %s", results[0].Name)
	}

	// Use FindAndCount with a filter that matches no documents
	filter = bson.M{"Name": "Nonexistent Model"}
	results = []sampleModel{}
	count, err = collection.FindAndCount(filter, &results, nil, ctx)
	if err != nil {
		t.Fatalf("Failed to find and count documents with non-matching filter: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count to be 0, got %d", count)
	}
	if len(results) != 0 {
		t.Fatalf("Expected 0 documents, got %d", len(results))
	}
}
func TestCollectionAll(t *testing.T) {
	ctx := context.WithValue(context.Background(), DbNameKey, "test-1")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB", "test")
	t.Setenv("ENV", "test")

	// Create a new collection
	collection, err := NewCollection[sampleModel](sampleModelCollectionName, ctx)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Clear the collection to ensure a clean slate
	err = collection.Clear(ctx)
	if err != nil {
		t.Fatalf("Failed to clear collection: %v", err)
	}

	// Add multiple documents to the collection
	model1 := collection.New()
	model1.Name = "Model 1"
	err = model1.Save(ctx)
	if err != nil {
		t.Fatalf("Failed to save model1: %v", err)
	}

	model2 := collection.New()
	model2.Name = "Model 2"
	err = model2.Save(ctx)
	if err != nil {
		t.Fatalf("Failed to save model2: %v", err)
	}

	// Use All to retrieve all documents
	var results []sampleModel
	err = collection.All(&results, nil, ctx)
	if err != nil {
		t.Fatalf("Failed to retrieve all documents: %v", err)
	}

	// Verify the retrieved documents
	if len(results) != 2 {
		t.Fatalf("Expected 2 documents, got %d", len(results))
	}

	names := []string{results[0].Name, results[1].Name}
	if !(names[0] == "Model 1" && names[1] == "Model 2") && !(names[0] == "Model 2" && names[1] == "Model 1") {
		t.Errorf("Expected documents with names 'Model 1' and 'Model 2', got %v", names)
	}
}
