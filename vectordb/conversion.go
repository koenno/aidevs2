package vectordb

import (
	"fmt"
	"reflect"

	qdrant "github.com/qdrant/go-client/qdrant"
)

const (
	tag = "qdrant"
)

func Marshal(item any) (*qdrant.PointStruct, error) {
	result := &qdrant.PointStruct{}
	payload := map[string]*qdrant.Value{}
	t := reflect.TypeOf(item)
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("item should be a struct, not a %T", item)
	}

	value := reflect.ValueOf(item)

	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		fieldValue := value.Field(i)

		tagVal := field.Tag.Get(tag)
		switch tagVal {
		case "":
			continue
		case "_id":
			result.Id = &qdrant.PointId{
				PointIdOptions: &qdrant.PointId_Uuid{
					Uuid: fieldValue.String(),
				},
			}
		case "_vector":
			slice, ok := fieldValue.Interface().([]float32)
			if !ok {
				return nil, fmt.Errorf("_vector should be of type []float32")
			}
			result.Vectors = &qdrant.Vectors{
				VectorsOptions: &qdrant.Vectors_Vector{
					Vector: &qdrant.Vector{
						Data: slice,
					},
				},
			}
		default:
			if field.Type.Kind() != reflect.String {
				return nil, fmt.Errorf("item payload field should be a string, not a %s", field.Type)
			}
			payload[tagVal] = &qdrant.Value{
				Kind: &qdrant.Value_StringValue{
					StringValue: fieldValue.String(),
				},
			}
		}
	}
	result.Payload = payload
	return result, nil
}

func UnmarshalScoredPoint(marshalled *qdrant.ScoredPoint, item any) error {
	t := reflect.TypeOf(item)
	if t.Kind() != reflect.Pointer {
		return fmt.Errorf("item should be a pointer to struct, not a %T", item)
	}

	value := reflect.ValueOf(item)
	value = value.Elem()

	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		fieldValue := value.Field(i)

		tagVal := field.Tag.Get(tag)
		switch tagVal {
		case "":
			continue
		case "_id":
			fieldValue.SetString(marshalled.Id.GetUuid())
		case "_vector":
			_, ok := fieldValue.Interface().([]float32)
			if !ok {
				return fmt.Errorf("_vector should be of type []float32")
			}
			fieldValue.Set(reflect.ValueOf(marshalled.Vectors.GetVector().GetData()))
		default:
			if field.Type.Kind() != reflect.String {
				return fmt.Errorf("item payload field should be a string, not a %s", field.Type)
			}
			payloadVal, exist := marshalled.Payload[tagVal]
			if !exist {
				continue
			}
			fieldValue.SetString(payloadVal.GetStringValue())
		}
	}
	return nil
}

func UnmarshalScoredPoints(marshalled []*qdrant.ScoredPoint, items any) error {
	t := reflect.TypeOf(items)
	if t.Kind() != reflect.Pointer {
		return fmt.Errorf("item should be a pointer to slice, not a %T", items)
	}

	typ := t.Elem()
	if typ.Kind() != reflect.Slice {
		return fmt.Errorf("items should be a pointer to a slice, not a %T", items)
	}

	item := typ.Elem()
	slice := reflect.MakeSlice(reflect.SliceOf(item), len(marshalled), len(marshalled))

	for i, scoredPoint := range marshalled {
		v := slice.Index(i).Addr().Interface()
		if err := UnmarshalScoredPoint(scoredPoint, v); err != nil {
			return fmt.Errorf("failed to unmarshal scored point: %v", err)
		}
	}
	value := reflect.ValueOf(items)
	value.Elem().Set(slice)
	return nil
}
