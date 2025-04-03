package bark_test

import (
	"context"
	"testing"

	"github.com/jaredtmartin/bark-go-mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestNewCollection(t *testing.T) {
	// Test creating collection with empty name
	emptyCollection := bark.NewCollection[*Dog]("")
	if emptyCollection.Name != "" {
		t.Errorf("Expected empty collection name, got %s", emptyCollection.Name)
	}

	// Test creating collection with valid name
	collectionName := "test_dogs"
	collection := bark.NewCollection[*Dog](collectionName)
	if collection.Name != collectionName {
		t.Errorf("Expected collection name %s, got %s", collectionName, collection.Name)
	}

	// Test that new collections are unique instances
	collection1 := bark.NewCollection[*Dog]("collection1")
	collection2 := bark.NewCollection[*Dog]("collection2")
	if collection1 == collection2 {
		t.Error("Expected different collection instances")
	}
	if collection1.Name == collection2.Name {
		t.Error("Expected different collection names")
	}
}
func TestMongoCollection(t *testing.T) {
	ctx := setupTest("MongoCollection", "2024-03-27T19:55:38.782Z", t)
	dogs := bark.NewCollection[*Dog]("dogs")

	t.Run("Test getting mongo collection first time", func(t *testing.T) {
		mongoCol1, err := dogs.MongoCollection(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if mongoCol1 == nil {
			t.Error("Expected non-nil mongo collection")
		}
	})

	// Test getting mongo collection first time

	t.Run("Test getting mongo collection second time (cached)", func(t *testing.T) {
		mongoCol1, err := dogs.MongoCollection(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		// Test getting mongo collection second time (cached)
		mongoCol2, err := dogs.MongoCollection(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if mongoCol2 == nil {
			t.Error("Expected non-nil mongo collection")
		}
		if mongoCol1 != mongoCol2 {
			t.Error("Expected same mongo collection instance on second call")
		}
	})
	t.Run("Test handling error when collection name is empty", func(t *testing.T) {
		emptyCollection := bark.NewCollection[*Dog]("")
		_, err := emptyCollection.MongoCollection(ctx)
		if err == nil || err.Error() != "collection name is required" {
			t.Errorf("Expected error 'collection name is required', got %v", err)
		}
	})
	t.Run("Test handling error when database is not set", func(t *testing.T) {
		ctx = context.WithValue(ctx, bark.MockDbErrorKey, "Mocked error")
		emptyDogs := bark.NewCollection[*Dog]("dogs")
		_, err := emptyDogs.MongoCollection(ctx)
		if err == nil {
			t.Error("Expected error when env var not set, got nil")
		}
	})
}
func TestFind(t *testing.T) {
	ctx := setupTest("Find", "2024-03-27T19:55:38.782Z", t)
	dogs := bark.NewCollection[*Dog]("dogs")
	SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111", Age: 3},
		{Name: "Spot", Id: "2222", Age: 5},
	}, ctx)
	t.Run("Find with valid filter", func(t *testing.T) {
		filter := bson.M{"Age": 3}
		results, err := dogs.Find(filter, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Name != "Fido" {
			t.Errorf("Expected dog name Fido, got %s", results[0].Name)
		}
	})

	t.Run("Find with empty filter", func(t *testing.T) {
		filter := bson.M{}
		results, err := dogs.Find(filter, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("Find with non-matching filter", func(t *testing.T) {
		filter := bson.M{"age": 99}
		results, err := dogs.Find(filter, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})

	t.Run("Find with options", func(t *testing.T) {
		opts := options.Find().SetLimit(1)
		results, err := dogs.Find(bson.M{}, opts, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result due to limit, got %d", len(results))
		}
	})

	t.Run("Find with invalid collection", func(t *testing.T) {
		invalidDogs := bark.NewCollection[*Dog]("")
		_, err := invalidDogs.Find(bson.M{}, nil, ctx)
		if err == nil {
			t.Error("Expected error for invalid collection, got nil")
		}
	})
	t.Run("sets collection names on each object returned", func(t *testing.T) {
		results, err := dogs.Find(bson.M{}, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		for _, dog := range results {
			if dog.CollectionName != "dogs" {
				t.Errorf("Expected collection name 'dogs', got %s", dog.CollectionName)
			}
		}
	})
}
func TestFindOne(t *testing.T) {
	ctx := setupTest("FindOne", "2024-03-27T19:55:38.782Z", t)
	dogs := bark.NewCollection[*Dog]("dogs")
	SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111", Age: 3},
		{Name: "Spot", Id: "2222", Age: 5},
	}, ctx)

	t.Run("FindOne with exact match", func(t *testing.T) {
		filter := bson.M{"Name": "Fido"}
		result, err := dogs.FindOne(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result.Name != "Fido" || result.Age != 3 {
			t.Errorf("Expected Fido with age 3, got %s with age %d", result.Name, result.Age)
		}
	})

	t.Run("FindOne with non-existent document", func(t *testing.T) {
		filter := bson.M{"Name": "NonExistent"}
		_, err := dogs.FindOne(filter, ctx)
		if err == nil {
			t.Error("Expected error for non-existent document, got nil")
		}
	})

	t.Run("FindOne with invalid collection name", func(t *testing.T) {
		invalidDogs := bark.NewCollection[*Dog]("")
		_, err := invalidDogs.FindOne(bson.M{}, ctx)
		if err == nil {
			t.Error("Expected error for invalid collection, got nil")
		}
	})

	t.Run("FindOne with complex filter", func(t *testing.T) {
		filter := bson.M{
			"Age":  bson.M{"$gt": 4},
			"Name": "Spot",
		}
		result, err := dogs.FindOne(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result.Name != "Spot" || result.Age != 5 {
			t.Errorf("Expected Spot with age 5, got %s with age %d", result.Name, result.Age)
		}
	})

	t.Run("FindOne with nil filter", func(t *testing.T) {
		result, err := dogs.FindOne(nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Error("Expected a result with nil filter, got nil")
		}
	})
}
func TestGet(t *testing.T) {
	ctx := setupTest("Get", "2024-03-27T19:55:38.782Z", t)
	dogs := bark.NewCollection[*Dog]("dogs")
	SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111", Age: 3},
		{Name: "Spot", Id: "2222", Age: 5},
	}, ctx)

	t.Run("Get with valid ID", func(t *testing.T) {
		result, err := dogs.Get("1111", ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result.Name != "Fido" || result.Age != 3 {
			t.Errorf("Expected Fido with age 3, got %s with age %d", result.Name, result.Age)
		}
	})

	t.Run("Get with non-existent ID", func(t *testing.T) {
		_, err := dogs.Get("9999", ctx)
		if err == nil {
			t.Error("Expected error for non-existent ID, got nil")
		}
	})

	t.Run("Get with empty ID", func(t *testing.T) {
		_, err := dogs.Get("", ctx)
		if err == nil {
			t.Error("Expected error for empty ID, got nil")
		}
	})

	t.Run("Get with invalid collection name", func(t *testing.T) {
		invalidDogs := bark.NewCollection[*Dog]("")
		_, err := invalidDogs.Get("1111", ctx)
		if err == nil {
			t.Error("Expected error for invalid collection, got nil")
		}
	})

	t.Run("Get verifies collection name is set", func(t *testing.T) {
		result, err := dogs.Get("2222", ctx)
		// fmt.Println("result: ", result)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result.CollectionName != "dogs" {
			t.Errorf("Expected collection name 'dogs', got %s", result.CollectionName)
		}
	})
}
func TestCount(t *testing.T) {
	ctx := setupTest("Count", "2024-03-27T19:55:38.782Z", t)
	dogs := bark.NewCollection[*Dog]("dogs")
	SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111", Age: 3},
		{Name: "Spot", Id: "2222", Age: 5},
		{Name: "Rex", Id: "3333", Age: 3},
	}, ctx)

	t.Run("Count with matching filter", func(t *testing.T) {
		filter := bson.M{"Age": 3}
		count, err := dogs.Count(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count of 2, got %d", count)
		}
	})

	t.Run("Count with empty filter", func(t *testing.T) {
		filter := bson.M{}
		count, err := dogs.Count(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count of 3, got %d", count)
		}
	})

	t.Run("Count with non-matching filter", func(t *testing.T) {
		filter := bson.M{"Age": 99}
		count, err := dogs.Count(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count of 0, got %d", count)
		}
	})

	t.Run("Count with complex filter", func(t *testing.T) {
		filter := bson.M{"Age": bson.M{"$gt": 3}}
		count, err := dogs.Count(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count of 1, got %d", count)
		}
	})

	t.Run("Count with invalid collection", func(t *testing.T) {
		invalidDogs := bark.NewCollection[*Dog]("")
		_, err := invalidDogs.Count(bson.M{}, ctx)
		if err == nil {
			t.Error("Expected error for invalid collection, got nil")
		}
	})

	t.Run("Count with nil filter", func(t *testing.T) {
		count, err := dogs.Count(nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count of 3, got %d", count)
		}
	})
}
func TestFindAndCount(t *testing.T) {
	ctx := setupTest("FindAndCount", "2024-03-27T19:55:38.782Z", t)
	dogs := bark.NewCollection[*Dog]("dogs")
	SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111", Age: 3},
		{Name: "Spot", Id: "2222", Age: 5},
		{Name: "Rex", Id: "3333", Age: 3},
		{Name: "Max", Id: "4444", Age: 7},
	}, ctx)

	t.Run("FindAndCount with age filter", func(t *testing.T) {
		filter := bson.M{"Age": 3}
		results, count, err := dogs.FindAndCount(filter, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count of 2, got %d", count)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		for _, dog := range results {
			if dog.Age != 3 {
				t.Errorf("Expected age 3, got %d", dog.Age)
			}
		}
	})

	t.Run("FindAndCount with pagination options", func(t *testing.T) {
		opts := options.Find().SetLimit(2).SetSkip(1)
		results, count, err := dogs.FindAndCount(bson.M{}, opts, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 4 {
			t.Errorf("Expected total count of 4, got %d", count)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results due to limit, got %d", len(results))
		}
	})

	t.Run("FindAndCount with complex filter", func(t *testing.T) {
		filter := bson.M{
			"Age":  bson.M{"$gt": 4},
			"Name": bson.M{"$in": []string{"Spot", "Max"}},
		}
		results, count, err := dogs.FindAndCount(filter, nil, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count of 2, got %d", count)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("FindAndCount with invalid collection", func(t *testing.T) {
		invalidDogs := bark.NewCollection[*Dog]("")
		_, _, err := invalidDogs.FindAndCount(bson.M{}, nil, ctx)
		if err == nil {
			t.Error("Expected error for invalid collection, got nil")
		}
	})

	t.Run("FindAndCount with sort options", func(t *testing.T) {
		opts := options.Find().SetSort(bson.M{"Age": -1}).SetLimit(2)
		results, count, err := dogs.FindAndCount(bson.M{}, opts, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 4 {
			t.Errorf("Expected total count of 4, got %d", count)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		if results[0].Age < results[1].Age {
			t.Error("Expected results to be sorted by age in descending order")
		}
	})
}
func TestAll(t *testing.T) {
	ctx := setupTest("All", "2024-03-27T19:55:38.782Z", t)
	dogs := bark.NewCollection[*Dog]("dogs")
	SetupFixture([]*Obj{
		{Name: "Fido", Id: "1111", Age: 3},
		{Name: "Spot", Id: "2222", Age: 5},
		{Name: "Rex", Id: "3333", Age: 3},
		{Name: "Max", Id: "4444", Age: 7},
	}, ctx)

	t.Run("All returns all documents", func(t *testing.T) {
		results, err := dogs.All(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 4 {
			t.Errorf("Expected 4 results, got %d", len(results))
		}
	})

	t.Run("All with empty collection", func(t *testing.T) {
		emptyDogs := bark.NewCollection[*Dog]("empty_dogs")
		results, err := emptyDogs.All(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})

	t.Run("All with invalid collection", func(t *testing.T) {
		invalidDogs := bark.NewCollection[*Dog]("")
		_, err := invalidDogs.All(ctx)
		if err == nil {
			t.Error("Expected error for invalid collection, got nil")
		}
	})

	t.Run("All verifies collection names", func(t *testing.T) {
		results, err := dogs.All(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		for _, dog := range results {
			if dog.CollectionName != "dogs" {
				t.Errorf("Expected collection name 'dogs', got %s", dog.CollectionName)
			}
		}
	})
}
func TestDeleteOne(t *testing.T) {
	ctx := setupTest("DeleteOne", "2024-03-27T19:55:38.782Z", t)
	dogs := bark.NewCollection[*Dog]("dogs")

	t.Run("DeleteOne with matching filter", func(t *testing.T) {
		SetupFixture([]*Obj{
			{Name: "Fido", Id: "1111", Age: 3},
			{Name: "Spot", Id: "2222", Age: 5},
			{Name: "Rex", Id: "3333", Age: 3},
		}, ctx)
		filter := bson.M{"Name": "Fido"}
		res, err := dogs.DeleteOne(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if res.Deleted != 1 {
			t.Errorf("Expected 1 deleted document, got %d", res.Deleted)
		}

		// Verify deletion
		count, err := dogs.Count(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error checking count, got %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count of 0 after deletion, got %d", count)
		}
	})

	t.Run("DeleteOne with non-matching filter", func(t *testing.T) {
		SetupFixture([]*Obj{
			{Name: "Fido", Id: "1111", Age: 3},
			{Name: "Spot", Id: "2222", Age: 5},
			{Name: "Rex", Id: "3333", Age: 3},
		}, ctx)
		filter := bson.M{"Name": "NonExistent"}
		res, err := dogs.DeleteOne(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error for non-matching filter, got %v", err)
		}
		if res.Deleted != 0 {
			t.Errorf("Expected 0 deleted documents, got %d", res.Deleted)
		}

		// Verify collection size unchanged
		count, err := dogs.Count(bson.M{}, ctx)
		if err != nil {
			t.Errorf("Expected no error checking count, got %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count of 3, got %d", count)
		}
	})

	t.Run("DeleteOne with complex filter", func(t *testing.T) {
		SetupFixture([]*Obj{
			{Name: "Fido", Id: "1111", Age: 3},
			{Name: "Spot", Id: "2222", Age: 5},
			{Name: "Rex", Id: "3333", Age: 3},
		}, ctx)
		filter := bson.M{
			"Age":  bson.M{"$gt": 4},
			"Name": "Spot",
		}
		res, err := dogs.DeleteOne(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if res.Deleted != 1 {
			t.Errorf("Expected 1 deleted document, got %d", res.Deleted)
		}
		// Verify specific document deleted
		result, err := dogs.FindOne(filter, ctx)
		if err == nil {
			t.Error("Expected error finding deleted document, got nil")
		}
		if result != nil {
			t.Error("Expected nil result for deleted document")
		}
	})

	t.Run("DeleteOne with invalid collection", func(t *testing.T) {
		invalidDogs := bark.NewCollection[*Dog]("")
		res, err := invalidDogs.DeleteOne(bson.M{}, ctx)
		if err == nil {
			t.Error("Expected error for invalid collection, got nil")
		}
		if res.Deleted != 0 {
			t.Error("Expected no deletions result for invalid collection")
		}
	})

	t.Run("DeleteOne with nil filter", func(t *testing.T) {
		SetupFixture([]*Obj{
			{Name: "Fido", Id: "1111", Age: 3},
			{Name: "Spot", Id: "2222", Age: 5},
			{Name: "Rex", Id: "3333", Age: 3},
		}, ctx)

		res, err := dogs.DeleteOne(nil, ctx)
		if err != nil {
			t.Errorf("Expected no error with nil filter, got %v", err)
		}
		if res.Deleted != 1 {
			t.Errorf("Expected 1 deleted documents, got %d", res.Deleted)
		}

		// Verify one document was deleted
		count, err := dogs.Count(bson.M{}, ctx)
		if err != nil {
			t.Errorf("Expected no error checking count, got %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count of 2 after deletion, got %d", count)
		}
	})
}
func TestDeleteMany(t *testing.T) {
	ctx := setupTest("DeleteMany", "2024-03-27T19:55:38.782Z", t)
	dogs := bark.NewCollection[*Dog]("dogs")

	t.Run("DeleteMany with filter", func(t *testing.T) {
		SetupFixture([]*Obj{
			{Name: "Fido", Id: "1111", Age: 3},
			{Name: "Spot", Id: "2222", Age: 5},
			{Name: "Rex", Id: "3333", Age: 3},
			{Name: "Max", Id: "4444", Age: 7},
		}, ctx)

		filter := bson.M{"Age": 3}
		res, err := dogs.DeleteMany(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if res.Deleted != 2 {
			t.Errorf("Expected 2 deleted documents, got %d", res.Deleted)
		}

		count, err := dogs.Count(bson.M{}, ctx)
		if err != nil {
			t.Errorf("Expected no error checking count, got %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count of 2 after deletion, got %d", count)
		}
	})

	t.Run("DeleteMany with range filter", func(t *testing.T) {
		SetupFixture([]*Obj{
			{Name: "Fido", Id: "1111", Age: 3},
			{Name: "Spot", Id: "2222", Age: 5},
			{Name: "Rex", Id: "3333", Age: 3},
			{Name: "Max", Id: "4444", Age: 7},
		}, ctx)

		filter := bson.M{"Age": bson.M{"$gte": 5}}
		res, err := dogs.DeleteMany(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if res.Deleted != 2 {
			t.Errorf("Expected 2 deleted documents, got %d", res.Deleted)
		}
	})

	t.Run("DeleteMany with no matching documents", func(t *testing.T) {
		SetupFixture([]*Obj{
			{Name: "Fido", Id: "1111", Age: 3},
			{Name: "Spot", Id: "2222", Age: 5},
		}, ctx)

		filter := bson.M{"Age": 99}
		res, err := dogs.DeleteMany(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if res.Deleted != 0 {
			t.Errorf("Expected 0 deleted documents, got %d", res.Deleted)
		}

		count, err := dogs.Count(bson.M{}, ctx)
		if err != nil {
			t.Errorf("Expected no error checking count, got %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count to remain 2, got %d", count)
		}
	})

	t.Run("DeleteMany with multiple field conditions", func(t *testing.T) {
		SetupFixture([]*Obj{
			{Name: "Fido", Id: "1111", Age: 3},
			{Name: "Spot", Id: "2222", Age: 5},
			{Name: "Rex", Id: "3333", Age: 3},
			{Name: "Max", Id: "4444", Age: 7},
		}, ctx)

		filter := bson.M{
			"Age":  bson.M{"$lt": 6},
			"Name": bson.M{"$in": []string{"Fido", "Spot"}},
		}
		res, err := dogs.DeleteMany(filter, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if res.Deleted != 2 {
			t.Errorf("Expected 2 deleted documents, got %d", res.Deleted)
		}
	})

	t.Run("DeleteMany with empty filter (delete all)", func(t *testing.T) {
		SetupFixture([]*Obj{
			{Name: "Fido", Id: "1111", Age: 3},
			{Name: "Spot", Id: "2222", Age: 5},
			{Name: "Rex", Id: "3333", Age: 3},
		}, ctx)

		res, err := dogs.DeleteMany(bson.M{}, ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if res.Deleted != 3 {
			t.Errorf("Expected 3 deleted documents, got %d", res.Deleted)
		}

		count, err := dogs.Count(bson.M{}, ctx)
		if err != nil {
			t.Errorf("Expected no error checking count, got %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count of 0 after deletion, got %d", count)
		}
	})
	t.Run("DeleteMany handling error getting collection", func(t *testing.T) {
		ctx = context.WithValue(ctx, bark.MockDbErrorKey, "Mocked error")
		_, err := dogs.DeleteMany(bson.M{}, ctx)
		if err == nil {
			t.Error("Expected error getting collection, got nil")
		}
	})
}
