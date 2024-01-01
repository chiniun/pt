package db

import (
	"context"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/reflect/protoreflect"

	"gorm.io/gorm/schema"
)

func init() {
	schema.RegisterSerializer("enum", ProtoEnumSerializer{})
}

// ProtoEnumSerializer proto enum serializer
type ProtoEnumSerializer struct {
}

// Scan implements serializer interface
func (ProtoEnumSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	fieldValue := reflect.New(field.FieldType)

	if dbValue != nil {
		var bytes []byte
		switch v := dbValue.(type) {
		case []byte:
			bytes = v
		case string:
			bytes = []byte(v)
		default:
			return fmt.Errorf("failed to unmarshal protocol enum value: %#v", dbValue)
		}

		v := fieldValue.Elem().Interface()
		nv, ok := v.(protoreflect.Enum)
		if !ok {
			return fmt.Errorf("type is not protocol enum")
		}

		for i := 0; i < nv.Descriptor().Values().Len(); i++ {
			if string(nv.Descriptor().Values().Get(i).Name()) == string(bytes) {
				newVal := nv.Type().New(protoreflect.EnumNumber(i))
				field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf(newVal))
				break
			}
		}
	}

	return nil
}

// Value implements serializer interface
func (ProtoEnumSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	nv, ok := fieldValue.(protoreflect.Enum)
	if !ok {
		return nil, fmt.Errorf("type is not protocol enum")
	}
	return string(nv.Descriptor().Values().Get(int(nv.Number())).Name()), nil
}
