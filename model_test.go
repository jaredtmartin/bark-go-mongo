package bark_test

import (
	"fmt"
	"testing"

	"github.com/jaredtmartin/bark-go-mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestSaveAndFetch(t *testing.T) {
	ctx := setupTest("SaveAndFetch", "2024-03-27T19:55:38.782Z", t)
	model, err := SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111"},
		{Name: "Spot", Id: "2222"},
	}, ctx)
	if err != nil {
		t.Fatalf("Failed to save fido: %v", err)
	}
	copy, err := model.Get("1111", ctx)
	if err != nil {
		t.Fatalf("Failed to get obj from db: %v", err)
	}
	if copy.Name != "Fido" {
		t.Fatalf("Expected name to be Fido, got %s", copy.Name)
	}
}
func TestGetAndSetId(t *testing.T) {
	fido := NewSampleModel("Fido")
	fido.SetId("1111")
	if fido.GetId() != "1111" {
		t.Fatalf("Expected id to be 1111, got %s", fido.GetId())
	}
	fido.SetId("2222")
	if fido.GetId() != "2222" {
		t.Fatalf("Expected id to be 2222, got %s", fido.GetId())
	}
}
func TestCollection(t *testing.T) {
	ctx := setupTest("Collection", "2024-03-27T19:55:38.782Z", t)

	// Test case: CollectionName is not set
	model := &sampleModel{
		DefaultModel: bark.DefaultModel[sampleModel]{},
	}
	_, err := model.Collection(ctx)
	if err == nil || err.Error() != "CollectionName not set" {
		t.Fatalf("Expected error 'CollectionName not set', got %v", err)
	}

	// Test case: Valid CollectionName
	// model.CollectionName = "test_collection"
	model = &sampleModel{
		DefaultModel: bark.DefaultModel[sampleModel]{
			CollectionName: "test_collection",
		},
	}

	collection, err := model.Collection(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if collection == nil {
		t.Fatalf("Expected collection to be non-nil")
	}

	// Test case: Cached collection
	cachedCollection, err := model.Collection(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if cachedCollection != collection {
		t.Fatalf("Expected cached collection to be the same as the first collection")
	}
}
func TestFind(t *testing.T) {
	ctx := setupTest("Find", "2024-03-27T19:55:38.782Z", t)
	model, err := SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111"},
		{Name: "Spot", Id: "2222"},
	}, ctx)
	if err != nil {
		t.Fatalf("Failed to save fido: %v", err)
	}
	fmt.Println(" fixture saved successfully")
	// Test case: Find documents with a filter
	filter := bson.M{"Name": "Fido"}
	opts := options.Find()
	results, err := model.Find(filter, opts, ctx)
	if err != nil {
		t.Fatalf("Failed to find documents: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Fido" {
		t.Fatalf("Expected Name to be Fido, got %s", results[0].Name)
	}

	// Test case: Find all documents
	filter = bson.M{}
	results, err = model.Find(filter, opts, ctx)
	if err != nil {
		t.Fatalf("Failed to find documents: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
}
func TestFindOne(t *testing.T) {
	ctx := setupTest("FindOne", "2024-03-27T19:55:38.782Z", t)

	// Setup: Insert sample data into the collection
	model, err := SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111"},
		{Name: "Spot", Id: "2222"},
	}, ctx)
	if err != nil {
		t.Fatalf("Failed to save fido: %v", err)
	}

	// Test case: Find a document with a valid filter
	filter := bson.M{"Id": "1111"}
	result, err := model.FindOne(filter, ctx)
	if err != nil {
		t.Fatalf("Failed to find document: %v", err)
	}
	if result == nil {
		t.Fatalf("Expected result to be non-nil")
	}
	if result.Name != "Fido" {
		t.Fatalf("Expected Name to be Fido, got %s", result.Name)
	}

	// Test case: Find a document with a filter that matches no documents
	filter = bson.M{"Id": "9999"}
	result, err = model.FindOne(filter, ctx)
	if err == nil || result != nil {
		t.Fatalf("Expected error or nil result for non-existent document, got result: %v, error: %v", result, err)
	}

	// Test case: Error when collection is not set
	model = &sampleModel{
		DefaultModel: bark.DefaultModel[sampleModel]{},
	}
	filter = bson.M{"Id": "1111"}
	_, err = model.FindOne(filter, ctx)
	if err == nil || err.Error() != "failed to get collection to save model to: CollectionName not set" {
		t.Fatalf("Expected error 'CollectionName not set', got %v", err)
	}
}
func TestGet(t *testing.T) {
	ctx := setupTest("Get", "2024-03-27T19:55:38.782Z", t)

	// Setup: Insert sample data into the collection
	model, err := SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111"},
		{Name: "Spot", Id: "2222"},
	}, ctx)
	if err != nil {
		t.Fatalf("Failed to save fido: %v", err)
	}

	// Test case: Get a document with a valid ID
	result, err := model.Get("1111", ctx)
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	if result == nil {
		t.Fatalf("Expected result to be non-nil")
	}
	if result.Name != "Fido" {
		t.Fatalf("Expected Name to be Fido, got %s", result.Name)
	}

	// Test case: Get a document with a non-existent ID
	result, err = model.Get("9999", ctx)
	if err == nil || result != nil {
		t.Fatalf("Expected error or nil result for non-existent document, got result: %v, error: %v", result, err)
	}

	// Test case: Error when collection is not set
	model = &sampleModel{
		DefaultModel: bark.DefaultModel[sampleModel]{},
	}
	_, err = model.Get("1111", ctx)
	if err == nil || err.Error() != "failed to get collection to save model to: CollectionName not set" {
		t.Fatalf("Expected error 'CollectionName not set', got %v", err)
	}
}
func TestLoad(t *testing.T) {
	ctx := setupTest("Load", "2024-03-27T19:55:38.782Z", t)

	// Setup: Insert sample data into the collection
	model, err := SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111"},
		{Name: "Spot", Id: "2222"},
	}, ctx)
	if err != nil {
		t.Fatalf("Failed to save fixture: %v", err)
	}
	fmt.Println("model:", model)
	// Test case: Successfully load an existing document
	model.SetId("1111")
	// I couldnt get it to set the fields on the same object,
	// so instead it just returns the object and you can assign to the model
	model, err = model.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load document: %v", err)
	}
	if model.Name != "Fido" {
		t.Fatalf("Expected Name to be Fido, got %s", model.Name)
	}

	// Test case: Attempt to load a non-existent document
	model.SetId("9999")
	_, err = model.Load(ctx)
	if err != bark.ErrObjNotFound {
		t.Fatalf("Expected ErrObjNotFound, got %v", err)
	}

	// Test case: Error when collection is not set
	model = &sampleModel{
		DefaultModel: bark.DefaultModel[sampleModel]{},
	}
	model.SetId("1111")
	_, err = model.Load(ctx)
	if err == nil || err.Error() != "failed to get collection to save model to: CollectionName not set" {
		t.Fatalf("Expected error 'CollectionName not set', got %v", err)
	}
}
