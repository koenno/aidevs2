package nosqldb

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName = "aidevs2"
)

type DB struct {
	client *mongo.Client
}

func New(addr string) (*DB, error) {
	URL := fmt.Sprintf("mongodb://%s", addr)
	opts := options.Client().ApplyURI(URL)
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %v", err)
	}
	return &DB{
		client: client,
	}, nil
}

func (db *DB) Close() {
	if err := db.client.Disconnect(context.Background()); err != nil {
		log.Printf("failed to close connection to mongo: %v", err)
	}
}

func (db *DB) InsertOne(ctx context.Context, collectionName string, item any) error {
	coll := db.client.Database(dbName).Collection(collectionName)
	_, err := coll.InsertOne(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to insert item: %v", err)
	}
	return nil
}

func (db *DB) InsertMany(ctx context.Context, collectionName string, items []any) error {
	coll := db.client.Database(dbName).Collection(collectionName)
	_, err := coll.InsertMany(ctx, items)
	if err != nil {
		return fmt.Errorf("failed to insert items: %v", err)
	}
	return nil
}

type SearchOption func(*searchOptions)

func WithFilter(key, val string) SearchOption {
	return func(so *searchOptions) {
		if so.filter == nil {
			so.filter = bson.M{}
		}
		so.filter[key] = val
	}
}

type searchOptions struct {
	filter bson.M
}

func (db *DB) Search(ctx context.Context, collectionName string, items any, options ...SearchOption) error {
	opts := &searchOptions{}
	for _, o := range options {
		o(opts)
	}
	coll := db.client.Database(dbName).Collection(collectionName)
	cur, err := coll.Find(ctx, opts.filter)
	if err != nil {
		return fmt.Errorf("failed to search items with filter %v: %v", opts.filter, err)
	}
	if err := cur.All(ctx, items); err != nil {
		return fmt.Errorf("failed to decode items got from db: %v", err)
	}
	return nil
}

func (db *DB) CollectionExist(ctx context.Context, collectionName string) (bool, error) {
	collections, err := db.client.Database(dbName).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return false, fmt.Errorf("failed to list collections: %v", err)
	}

	for _, c := range collections {
		if c == collectionName {
			return true, nil
		}
	}
	return false, nil
}
