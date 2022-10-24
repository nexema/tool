package internal

import (
	"fmt"
	"reflect"
)

// Provides a method to check if an interface is convertible to a FieldTypePrimitive
type TypeChecker struct{}

func NewTypeChecker() *TypeChecker {
	return &TypeChecker{}
}

// ValidateType validates that v is convertible to type t
// if v is not of type t, an error is returned, otherwise, nil
// o contains the converted type if an int or uint
func (TypeChecker) ValidateType(v interface{}, t FieldTypePrimitive) (o interface{}, err error) {
	vR := reflect.TypeOf(v).Kind()

	success := true
	var newVal interface{}

	switch t {
	case Boolean:
		if vR != reflect.Bool {
			success = false
		}

	case String:
		if vR != reflect.String {
			success = false
		}

	case Uint8:
		out, ok := uintCaster[uint8](v, vR)
		success = ok
		newVal = out

	case Uint16:
		out, ok := uintCaster[uint16](v, vR)
		success = ok
		newVal = out

	case Uint32:
		out, ok := uintCaster[uint32](v, vR)
		success = ok
		newVal = out

	case Uint64:
		if vR != reflect.Uint64 {
			success = false
		}

	case Int8:
		out, ok := intCaster[int8](v, vR)
		success = ok
		newVal = out

	case Int16:
		out, ok := intCaster[int16](v, vR)
		success = ok
		newVal = out

	case Int32:
		out, ok := intCaster[int32](v, vR)
		success = ok
		newVal = out

	case Int64:
		if vR != reflect.Int64 {
			success = false
		}

	case Float32:
		if vR != reflect.Float64 {
			success = false
		}

		out, ok := safeNumberCaster[float64, float32](v.(float64))
		success = ok
		newVal = out

	case Float64:
		if vR != reflect.Float64 {
			success = false
		}
	}

	if success {
		return nil, nil
	}

	return newVal, fmt.Errorf("value %v is not of type %s", v, t)
}

func uintCaster[O numeric](in interface{}, r reflect.Kind) (out O, ok bool) {
	if r == reflect.Uint64 {
		return safeNumberCaster[uint64, O](in.(uint64))
	}

	if r == reflect.Int64 {
		return safeNumberCaster[int64, O](in.(int64))

	}
	return 0, false
}

func intCaster[O numeric](in interface{}, r reflect.Kind) (out O, ok bool) {
	if r != reflect.Int64 {
		return 0, false
	}
	return safeNumberCaster[int64, O](in.(int64))
}

func safeNumberCaster[I numeric, O numeric](in I) (out O, ok bool) {
	if I(O(in)) != in {
		return 0, false
	}

	return O(in), true
}

type numeric interface {
	int8 | int16 | int32 | int64 |
		uint8 | uint16 | uint32 | uint64 | float32 | float64
}
