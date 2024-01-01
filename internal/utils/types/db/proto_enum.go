package db

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

type ProtoEnum[T protoreflect.Enum] struct {
	data T
}

func NewProtoEnum[T protoreflect.Enum](data T) ProtoEnum[T] {
	return ProtoEnum[T]{
		data: data,
	}
}

func (p *ProtoEnum[T]) Scan(src any) error {
	var bytes []byte
	switch v := src.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("failed to unmarshal protocol enum value", src))
	}

	for i := 0; i < p.data.Descriptor().Values().Len(); i++ {
		numString := protoimpl.X.EnumStringOf(p.Data().Descriptor(), protoreflect.EnumNumber(i))
		if numString == string(bytes) {
			var nd ProtoEnum[protoreflect.Enum]
			nd.data = p.data.Type().New(protoreflect.EnumNumber(i))
			pr := reflect.ValueOf(&p.data).Elem()
			pr.Set(reflect.ValueOf(nd.data))
			break
		}

	}

	return nil
}

func (p ProtoEnum[T]) Value() (driver.Value, error) {
	return protoimpl.X.EnumStringOf(p.Data().Descriptor(), protoreflect.EnumNumber(p.Data().Number())), nil
}

func (p ProtoEnum[T]) Data() T {
	return p.data
}
