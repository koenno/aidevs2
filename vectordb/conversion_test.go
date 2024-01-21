package vectordb

import (
	"testing"

	qdrant "github.com/qdrant/go-client/qdrant"
	"github.com/stretchr/testify/assert"
)

type Item struct {
	ID     string    `qdrant:"_id"`
	Vector []float32 `qdrant:"_vector"`
	Name   string    `qdrant:"name"`
	URL    string    `qdrant:"url"`
}

func TestShouldConvertToQdrantFormat(t *testing.T) {
	// given
	data := Item{
		ID:     "e5b5c018-c511-4249-b139-ea4c9dc6668b",
		Vector: []float32{0.75, 0.01, 0.0, 1.0, 0.99},
		Name:   "some name",
		URL:    "http://some.url",
	}
	expectedData := &qdrant.PointStruct{
		Id: &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Uuid{
				Uuid: data.ID,
			},
		},
		Vectors: &qdrant.Vectors{
			VectorsOptions: &qdrant.Vectors_Vector{
				Vector: &qdrant.Vector{
					Data: data.Vector,
				},
			},
		},
		Payload: map[string]*qdrant.Value{
			"name": {
				Kind: &qdrant.Value_StringValue{
					StringValue: data.Name,
				},
			},
			"url": {
				Kind: &qdrant.Value_StringValue{
					StringValue: data.URL,
				},
			},
		},
	}

	// when
	marshalled, err := Marshal(data)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedData, marshalled)
}

func TestShouldUnmarshalScoredPoint(t *testing.T) {
	// given
	expectedData := Item{
		ID:     "e5b5c018-c511-4249-b139-ea4c9dc6668b",
		Vector: []float32{0.75, 0.01, 0.0, 1.0, 0.99},
		Name:   "some name",
		URL:    "http://some.url",
	}
	data := &qdrant.ScoredPoint{
		Id: &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Uuid{
				Uuid: expectedData.ID,
			},
		},
		Vectors: &qdrant.Vectors{
			VectorsOptions: &qdrant.Vectors_Vector{
				Vector: &qdrant.Vector{
					Data: expectedData.Vector,
				},
			},
		},
		Payload: map[string]*qdrant.Value{
			"name": {
				Kind: &qdrant.Value_StringValue{
					StringValue: expectedData.Name,
				},
			},
			"url": {
				Kind: &qdrant.Value_StringValue{
					StringValue: expectedData.URL,
				},
			},
		},
	}

	var unmarshalled Item

	// when
	err := UnmarshalScoredPoint(data, &unmarshalled)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedData, unmarshalled)
}

func TestShouldUnmarshalScoredPoints(t *testing.T) {
	// given
	expectedData := []Item{
		{
			ID:     "e5b5c018-c511-4249-b139-ea4c9dc6668b",
			Vector: []float32{0.75, 0.01, 0.0, 1.0, 0.99},
			Name:   "first name",
			URL:    "http://some.url",
		},
		{
			ID:     "f45ab456-7632-ef54-bc71-876cd321baf6",
			Vector: []float32{0.987654},
			Name:   "second name",
			URL:    "",
		},
	}
	data := []*qdrant.ScoredPoint{
		{
			Id: &qdrant.PointId{
				PointIdOptions: &qdrant.PointId_Uuid{
					Uuid: expectedData[0].ID,
				},
			},
			Vectors: &qdrant.Vectors{
				VectorsOptions: &qdrant.Vectors_Vector{
					Vector: &qdrant.Vector{
						Data: expectedData[0].Vector,
					},
				},
			},
			Payload: map[string]*qdrant.Value{
				"name": {
					Kind: &qdrant.Value_StringValue{
						StringValue: expectedData[0].Name,
					},
				},
				"url": {
					Kind: &qdrant.Value_StringValue{
						StringValue: expectedData[0].URL,
					},
				},
			},
		},
		{
			Id: &qdrant.PointId{
				PointIdOptions: &qdrant.PointId_Uuid{
					Uuid: expectedData[1].ID,
				},
			},
			Vectors: &qdrant.Vectors{
				VectorsOptions: &qdrant.Vectors_Vector{
					Vector: &qdrant.Vector{
						Data: expectedData[1].Vector,
					},
				},
			},
			Payload: map[string]*qdrant.Value{
				"name": {
					Kind: &qdrant.Value_StringValue{
						StringValue: expectedData[1].Name,
					},
				},
			},
		},
	}

	var unmarshalled []Item

	// when
	err := UnmarshalScoredPoints(data, &unmarshalled)

	// then
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedData, unmarshalled)
	assert.Equal(t, len(expectedData), len(unmarshalled))
}
