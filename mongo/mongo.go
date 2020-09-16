package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/seoulstore/price-crawler/search"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func MongoCTX() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func GetMongoClient() (*mongo.Client, error) {
	c, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://monkey:Qkskskdndb@sinsa-ustho.mongodb.net/sinsa-dev?retryWrites=true"))
	if err != nil {
		return nil, err
	}

	ctx, cancel := MongoCTX()
	defer cancel()
	err = c.Connect(ctx)
	if err != nil {
		return nil, err
	}

	client = c

	return c, nil

	// defer func() {
	// 	if err = client.Disconnect(ctx); err != nil {
	// 		panic(err)
	// 	}
	// }()

	// databases, err := client.ListDatabaseNames(ctx, bson.D{})
	// fmt.Println(databases)

	// collection, err := client.Database("sinsa-dev").Collection("site-product-edit-data").Find(ctx, bson.D{})
	// if err != nil {
	// 	panic(err)
	// }
	// var r interface{}
	// collection.All(ctx, r)
	// fmt.Println(r)
}

type query struct {
	ID    primitive.ObjectID `bson:"_id"`
	Query string             `bson:"query"`
}

type Result struct {
	ID              primitive.ObjectID `bson:"_id"`
	MallProducts    []*search.EP       `bson:"mallProducts"`
	CompareProducts []*search.CP       `bson:"compareProducts"`
}

func GetQueries() ([]query, error) {
	ctx, cancel := MongoCTX()
	defer cancel()

	db := client.Database("sinsa-dev").Collection("price-crawler")
	cur, err := db.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var results []query
	if err = cur.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func UpdateQuery(result *Result) error {
	ctx, cancel := MongoCTX()
	defer cancel()

	db := client.Database("sinsa-dev").Collection("price-crawler")
	r, err := db.UpdateOne(
		ctx,
		bson.M{"_id": result.ID},
		bson.D{
			{"$set", bson.D{{"mallProducts", result.MallProducts}, {"compareProducts", result.CompareProducts}}},
		},
	)

	fmt.Println(r)

	if err != nil {
		return err
	}

	return nil
}
