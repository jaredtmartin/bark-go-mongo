package bark_test

import (
	"testing"

	"github.com/jaredtmartin/bark-go-mongo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestSaveAndFetch(t *testing.T) {
	ctx := setupTest("SaveAndFetch", "2024-03-27T19:55:38.782Z", t)
	collection, err := SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111"},
		{Name: "Spot", Id: "2222"},
	}, ctx)
	if err != nil {
		t.Fatalf("Failed to save fido: %v", err)
	}
	copy, err := collection.Get("1111", ctx)
	if err != nil {
		t.Fatalf("Failed to get obj from db: %v", err)
	}
	if copy.Name != "Fido" {
		t.Fatalf("Expected name to be Fido, got %s", copy.Name)
	}
}
func TestCollection(t *testing.T) {
	ctx := setupTest("Collection", "2024-03-27T19:55:38.782Z", t)

	// Test case: CollectionName is not set
	model := &Dog{
		Model: bark.Model{},
	}
	_, err := model.Collection().MongoCollection(ctx)
	if err == nil || err.Error() != "collection name is required" {
		t.Fatalf("Expected error 'collection name is required', got %v", err)
	}

	// Test case: Valid CollectionName
	// model.CollectionName = "test_collection"
	model = &Dog{
		Model: bark.Model{
			CollectionName: "test_collection",
		},
	}

	collection, err := model.Collection().MongoCollection(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if collection == nil {
		t.Fatalf("Expected collection to be non-nil")
	}

	// Test case: Cached collection
	cachedCollection, err := model.Collection().MongoCollection(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if cachedCollection != collection {
		t.Fatalf("Expected cached collection to be the same as the first collection")
	}
}
func TestGetCollectionName(t *testing.T) {
	// Test case: CollectionName is set
	model := &bark.Model{
		CollectionName: "test_collection",
	}
	if model.GetCollectionName() != "test_collection" {
		t.Fatalf("Expected collection name to be 'test_collection', got '%s'", model.GetCollectionName())
	}

	// Test case: CollectionName is not set
	model = &bark.Model{}
	if model.GetCollectionName() != "" {
		t.Fatalf("Expected collection name to be empty, got '%s'", model.GetCollectionName())
	}
}
func TestDelete(t *testing.T) {
	ctx := setupTest("ModelDelete", "2024-03-27T19:55:38.782Z", t)

	// Test case: Delete with no ID set
	model := &bark.Model{CollectionName: "test_collection"}
	_, err := model.Delete(ctx)
	if err == nil || err.Error() != "cannot delete model with no id" {
		t.Fatalf("Expected error 'cannot delete model with no id', got %v", err)
	}

	// Test case: Delete with valid ID
	model = &bark.Model{
		ID:             "1234",
		Id:             "1234",
		CollectionName: "test_collection",
	}
	model.SaveModel(model, ctx)
	_, err = SetupFixture([]*Obj{
		{Id: "1234", Name: "TestObj"},
	}, ctx)
	if err != nil {
		t.Fatalf("Failed to set up fixture: %v", err)
	}

	result, err := model.Delete(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.Deleted != 1 {
		t.Fatalf("Expected 1 document to be deleted, got %d", result.Deleted)
	}

	// Test case: Delete non-existent ID
	model = &bark.Model{
		ID:             "5678",
		Id:             "5678",
		CollectionName: "test_collection",
	}
	result, err = model.Delete(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.Deleted != 0 {
		t.Fatalf("Expected 0 documents to be deleted, got %d", result.Deleted)
	}
}
func TestSaveModel(t *testing.T) {
	ctx := setupTest("SaveModel", "2024-03-27T19:55:38.782Z", t)
	spot := NewDog("Spot")
	spot.Id = "2222"
	_, err := SetupFixture([]*Obj{spot.ToFixture()}, ctx)
	if err != nil {
		t.Fatalf("Failed to set up fixture: %v", err)
	}
	t.Run("Save model with no ID (new object)", func(t *testing.T) {
		// Test case: Save model with no ID (new object)
		fido := NewDog("Fido")

		result, err := fido.SaveModel(fido, ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		// fmt.Println("insert result", result.String())
		if result.Inserted != 1 {
			t.Fatalf("Expected 1 document to be inserted, got %d", result.Inserted)
		}
		if fido.Id == "" || fido.ID == "" {
			t.Fatalf("Expected model ID to be set, got empty ID")
		}
	})
	t.Run("Save model with existing ID (update object)", func(t *testing.T) {
		// Test case: Save model with existing ID (update object)
		spot.Age = 5
		spot.Name = "Spotty"
		result, err := spot.SaveModel(spot, ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Matched != 1 {
			t.Fatalf("Expected 1 document to be matched, got %d", result.Matched)
		}
		if result.Modified != 1 {
			t.Fatalf("Expected 1 document to be modified, got %d", result.Modified)
		}
		if result.Inserted != 0 {
			t.Fatalf("Expected 0 documents to be inserted, got %d", result.Inserted)
		}
	})
	t.Run("Save model with invalid collection", func(t *testing.T) {
		// Test case: Save model with invalid collection
		model := &bark.Model{}
		_, err := model.SaveModel(model, ctx)
		if err == nil || err.Error() != "failed to get collection to save model to: collection name is required" {
			t.Fatalf("Expected error 'failed to get collection to save model to: collection name is required', got %v", err)
		}
	})
}
func TestEmptyResult(t *testing.T) {
	t.Run("Creates new empty result", func(t *testing.T) {
		result := bark.EmptyResult()
		if result == nil {
			t.Fatal("Expected non-nil Result")
		}
		if result.Matched != 0 {
			t.Errorf("Expected Matched to be 0, got %d", result.Matched)
		}
		if result.Modified != 0 {
			t.Errorf("Expected Modified to be 0, got %d", result.Modified)
		}
		if result.Deleted != 0 {
			t.Errorf("Expected Deleted to be 0, got %d", result.Deleted)
		}
		if result.Inserted != 0 {
			t.Errorf("Expected Inserted to be 0, got %d", result.Inserted)
		}
	})

	t.Run("Multiple calls return different instances", func(t *testing.T) {
		result1 := bark.EmptyResult()
		result2 := bark.EmptyResult()
		if result1 == result2 {
			t.Error("Expected different instances of Result")
		}
	})

	t.Run("Instance is mutable", func(t *testing.T) {
		result := bark.EmptyResult()
		result.Matched = 1
		result.Modified = 2
		result.Deleted = 3
		result.Inserted = 4

		if result.Matched != 1 || result.Modified != 2 || result.Deleted != 3 || result.Inserted != 4 {
			t.Error("Expected Result to be mutable")
		}
	})
}
func TestResultFromUpdate(t *testing.T) {
	t.Run("Creates result from non-nil update result", func(t *testing.T) {
		updateResult := &mongo.UpdateResult{
			MatchedCount:  5,
			ModifiedCount: 3,
			UpsertedCount: 1,
		}
		result := bark.ResultFromUpdate(updateResult)

		if result.Matched != 5 {
			t.Errorf("Expected Matched count to be 5, got %d", result.Matched)
		}
		if result.Modified != 3 {
			t.Errorf("Expected Modified count to be 3, got %d", result.Modified)
		}
		if result.Inserted != 1 {
			t.Errorf("Expected Inserted count to be 1, got %d", result.Inserted)
		}
		if result.Deleted != 0 {
			t.Errorf("Expected Deleted count to be 0, got %d", result.Deleted)
		}
	})

	t.Run("Creates result from zero-value update result", func(t *testing.T) {
		updateResult := &mongo.UpdateResult{}
		result := bark.ResultFromUpdate(updateResult)

		if result.Matched != 0 {
			t.Errorf("Expected Matched count to be 0, got %d", result.Matched)
		}
		if result.Modified != 0 {
			t.Errorf("Expected Modified count to be 0, got %d", result.Modified)
		}
		if result.Inserted != 0 {
			t.Errorf("Expected Inserted count to be 0, got %d", result.Inserted)
		}
		if result.Deleted != 0 {
			t.Errorf("Expected Deleted count to be 0, got %d", result.Deleted)
		}
	})

	t.Run("Preserves large number values", func(t *testing.T) {
		updateResult := &mongo.UpdateResult{
			MatchedCount:  999999,
			ModifiedCount: 888888,
			UpsertedCount: 777777,
		}
		result := bark.ResultFromUpdate(updateResult)

		if result.Matched != 999999 {
			t.Errorf("Expected Matched count to be 999999, got %d", result.Matched)
		}
		if result.Modified != 888888 {
			t.Errorf("Expected Modified count to be 888888, got %d", result.Modified)
		}
		if result.Inserted != 777777 {
			t.Errorf("Expected Inserted count to be 777777, got %d", result.Inserted)
		}
	})
}
func TestResultFromDelete(t *testing.T) {
	t.Run("Creates result from non-nil delete result", func(t *testing.T) {
		deleteResult := &mongo.DeleteResult{
			DeletedCount: 5,
		}
		result := bark.ResultFromDelete(deleteResult)

		if result.Deleted != 5 {
			t.Errorf("Expected Deleted count to be 5, got %d", result.Deleted)
		}
		if result.Matched != 0 {
			t.Errorf("Expected Matched count to be 0, got %d", result.Matched)
		}
		if result.Modified != 0 {
			t.Errorf("Expected Modified count to be 0, got %d", result.Modified)
		}
		if result.Inserted != 0 {
			t.Errorf("Expected Inserted count to be 0, got %d", result.Inserted)
		}
	})

	t.Run("Creates result from zero-value delete result", func(t *testing.T) {
		deleteResult := &mongo.DeleteResult{}
		result := bark.ResultFromDelete(deleteResult)

		if result.Deleted != 0 {
			t.Errorf("Expected Deleted count to be 0, got %d", result.Deleted)
		}
		if result.Matched != 0 {
			t.Errorf("Expected Matched count to be 0, got %d", result.Matched)
		}
		if result.Modified != 0 {
			t.Errorf("Expected Modified count to be 0, got %d", result.Modified)
		}
		if result.Inserted != 0 {
			t.Errorf("Expected Inserted count to be 0, got %d", result.Inserted)
		}
	})

	t.Run("Preserves large number values", func(t *testing.T) {
		deleteResult := &mongo.DeleteResult{
			DeletedCount: 999999,
		}
		result := bark.ResultFromDelete(deleteResult)

		if result.Deleted != 999999 {
			t.Errorf("Expected Deleted count to be 999999, got %d", result.Deleted)
		}
	})
}
func TestResultString(t *testing.T) {
	t.Run("String representation of empty result", func(t *testing.T) {
		result := bark.Result{}
		expected := "Matched: 0, Modified: 0, Inserted: 0, Deleted: 0"
		if result.String() != expected {
			t.Errorf("Expected %s, got %s", expected, result.String())
		}
	})

	t.Run("String representation with all fields populated", func(t *testing.T) {
		result := bark.Result{
			Matched:  42,
			Modified: 24,
			Inserted: 15,
			Deleted:  7,
		}
		expected := "Matched: 42, Modified: 24, Inserted: 15, Deleted: 7"
		if result.String() != expected {
			t.Errorf("Expected %s, got %s", expected, result.String())
		}
	})

	t.Run("String representation with max int64 values", func(t *testing.T) {
		result := bark.Result{
			Matched:  9223372036854775807,
			Modified: 9223372036854775807,
			Inserted: 9223372036854775807,
			Deleted:  9223372036854775807,
		}
		expected := "Matched: 9223372036854775807, Modified: 9223372036854775807, Inserted: 9223372036854775807, Deleted: 9223372036854775807"
		if result.String() != expected {
			t.Errorf("Expected %s, got %s", expected, result.String())
		}
	})

	t.Run("String representation with mixed values", func(t *testing.T) {
		result := bark.Result{
			Matched:  1,
			Modified: 0,
			Inserted: 5,
			Deleted:  0,
		}
		expected := "Matched: 1, Modified: 0, Inserted: 5, Deleted: 0"
		if result.String() != expected {
			t.Errorf("Expected %s, got %s", expected, result.String())
		}
	})
}
