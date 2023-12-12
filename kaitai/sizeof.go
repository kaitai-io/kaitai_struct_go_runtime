package kaitai

import (
	"errors"
	"reflect"
)

type Size interface {
	Size() (uint64, error)
}

func SizeOf(msg interface{}) (uint64, error) {
	v := reflect.ValueOf(msg)
	return sizeOf(v)
}

func sizeOf(fieldValue reflect.Value) (uint64, error) {
	// interface
	marshaler, ok := fieldValue.Interface().(Size)
	if ok && marshaler != nil {
		return marshaler.Size()
	}

	switch fieldValue.Kind() {
	case reflect.Uint8, reflect.Int8, reflect.Bool:
		return 1, nil
	case reflect.Uint16, reflect.Int16:
		return 2, nil
	case reflect.Uint32, reflect.Int32, reflect.Float32:
		return 4, nil
	case reflect.Uint64, reflect.Int64, reflect.Float64:
		return 8, nil
	case reflect.String:
		return uint64(len(fieldValue.String())), nil
	case reflect.Slice, reflect.Array:
		arrLen := fieldValue.Len()
		var result uint64
		for i := 0; i < arrLen; i++ {
			v := fieldValue.Index(i)
			vl, err := sizeOf(v)
			if err != nil {
				return 0, err
			}
			result += vl
		}
		return result, nil
	case reflect.Struct:
		numField := fieldValue.NumField()
		valueType := fieldValue.Type()
		var result uint64
		for i := 0; i < numField; i++ {
			fieldType := valueType.Field(i)
			// skip stream and not exported field
			if fieldType.Anonymous || fieldType.PkgPath != "" {
				continue
			}

			v := fieldValue.Field(i)
			vl, err := sizeOf(v)
			if err != nil {
				return 0, err
			}
			result += vl
		}
		return result, nil
	case reflect.Pointer:
		// nil 0
		if fieldValue.IsNil() {
			return 0, nil
		}
		fieldValue = fieldValue.Elem()
		return sizeOf(fieldValue)
	case reflect.Interface:
		if fieldValue.IsNil() {
			return 0, errors.New(fieldValue.Kind().String() + " interface nil")
		}
		fieldValue = fieldValue.Elem()
		return sizeOf(fieldValue)
	default:
		return 0, errors.New(`type "` + fieldValue.Kind().String() + `" not supported`)
	}
}