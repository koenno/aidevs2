package vectordb

import (
	"context"
	"fmt"
	"log"

	qdrant "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	VectorSize = 1536
	Distance   = qdrant.Distance_Cosine
)

type DB struct {
	conn *grpc.ClientConn
}

func New(addr string) (*DB, error) {
	db := &DB{}
	var err error
	db.conn, err = grpc.DialContext(context.Background(), addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to qdrant: %v", err)
	}

	return db, nil
}

func (db *DB) Close() {
	if err := db.conn.Close(); err != nil {
		log.Printf("failed to close connection to qdrant: %v", err)
	}
}

func (db *DB) CreateCollection(ctx context.Context, collectionName string) error {
	client := qdrant.NewCollectionsClient(db.conn)
	var defaultSegmentNumber uint64 = 2
	_, err := client.Create(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     VectorSize,
					Distance: Distance,
				},
			}},
		OptimizersConfig: &qdrant.OptimizersConfigDiff{
			DefaultSegmentNumber: &defaultSegmentNumber,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create collection '%s': %v", collectionName, err)
	}
	return nil
}

func (db *DB) UpsertOne(ctx context.Context, collectionName string, item any) error {
	data, err := Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %v", err)
	}
	waitUpsert := true
	client := qdrant.NewPointsClient(db.conn)
	_, err = client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &waitUpsert,
		Points:         []*qdrant.PointStruct{data},
	})
	if err != nil {
		return fmt.Errorf("failed to upsert vector: %v", err)
	}
	return nil
}

func (db *DB) UpsertMany(ctx context.Context, collectionName string, items []any) error {
	for _, item := range items {
		if err := db.UpsertOne(ctx, collectionName, item); err != nil {
			return fmt.Errorf("failed to upsert many: %v", err)
		}
	}
	return nil
}

type SearchOption func(*searchOptions)

func WithLimit(n uint64) SearchOption {
	return func(so *searchOptions) {
		so.limit = n
	}
}

type searchOptions struct {
	limit uint64
}

func (db *DB) Search(ctx context.Context, collectionName string, vector []float32, items any, options ...SearchOption) error {
	opts := &searchOptions{}
	for _, o := range options {
		o(opts)
	}
	client := qdrant.NewPointsClient(db.conn)
	res, err := client.Search(ctx, &qdrant.SearchPoints{
		CollectionName: collectionName,
		Vector:         vector,
		Limit:          opts.limit,
		// Include all payload and vectors in the search result
		WithVectors: &qdrant.WithVectorsSelector{SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true}},
		WithPayload: &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		return fmt.Errorf("failed to search vector: %v", err)
	}

	return UnmarshalScoredPoints(res.Result, items)
}

func (db *DB) CollectionExist(ctx context.Context, collectionName string) (bool, error) {
	client := qdrant.NewCollectionsClient(db.conn)
	res, err := client.List(ctx, &qdrant.ListCollectionsRequest{})
	if err != nil {
		return false, fmt.Errorf("failed to list collections: %v", err)
	}

	for _, c := range res.Collections {
		if c.Name == collectionName {
			return true, nil
		}
	}
	return false, nil
}
