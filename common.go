package bark

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrClearCanOnlyBeUsedOnDbsStartingWithTest = errors.New("to prevent accidents, clear method can only be used on databases whose names start with 'test'")

// Returns all documents matching the filter
func Find(collection *mongo.Collection, filter bson.M, results interface{}, opts *options.FindOptionsBuilder, ctx context.Context) error {
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return fmt.Errorf("error fetching documents: %v", err)
	}
	if err = cursor.All(ctx, results); err != nil {
		return fmt.Errorf("error decoding documents: %v", err)
	}
	return nil
}

// Returns the total number of documents matching the filter
func Count(collection *mongo.Collection, filter bson.M, ctx context.Context) (int64, error) {
	return collection.CountDocuments(ctx, filter)
}

// Returns the total number of documents matching the filter and returns the results
func FindAndCount(collection *mongo.Collection, filter bson.M, results interface{}, opts *options.FindOptionsBuilder, ctx context.Context) (int64, error) {
	err := Find(collection, filter, results, opts, ctx)
	if err != nil {
		return 0, err
	}
	count, err2 := Count(collection, filter, ctx)
	return count, err2
}

// Returns all documents in the collection
func All(collection *mongo.Collection, results interface{}, opts *options.FindOptionsBuilder, ctx context.Context) error {
	filter := bson.M{}
	return Find(collection, filter, results, opts, ctx)
}

// func Save(model Model, ctx context.Context) error {
// 	// We don't ask for the collection, because we need the model and ctx here anyway,
// 	// so we can get the collection ourselves.
// 	collection, err := model.Collection(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to get collection to save model to: %v", err)
// 	}
// 	id := model.GetId()
// 	model.SetId(id)
// 	fmt.Println("obj has id: ", id)
// 	if model.GetId() == "" {
// 		model.SetId(Uuid())
// 	}
// 	fmt.Println("obj will be saved with id: ", id)
// 	filter := bson.M{"Id": model.GetId()}
// 	update := bson.M{
// 		"$set": model,
// 		"$inc": bson.M{"Version": 1},
// 		"$setOnInsert": bson.M{
// 			"CreatedOn": Now(ctx),
// 		},
// 	}
// 	opts := options.UpdateOne().SetUpsert(true)
// 	res, err := collection.UpdateOne(ctx, filter, update, opts)
// 	fmt.Println("Saved: Matched:", res.MatchedCount, " Modified: ", res.ModifiedCount, " Upserted: ", res.UpsertedCount, " UpsertedID: ", res.UpsertedID)

//		if err != nil {
//			return err
//		}
//		return nil
//	}
// func Delete(model Model, ctx context.Context) error {
// 	// We don't ask for the collection, because we need the model and ctx here anyway,
// 	// so we can get the collection ourselves.
// 	fmt.Printf("Deleting %v\n", model)
// 	collection, err := model.Collection(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to get collection to save model to: %v", err)
// 	}
// 	filter := bson.M{"Id": model.GetId()}
// 	res, err := collection.DeleteOne(ctx, filter, nil)
// 	fmt.Println("Delete result: ", res)
// 	return err
// }
