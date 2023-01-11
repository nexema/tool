package internal

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
)

// hash hashes input using sha256
var hashInst hash.Hash = sha256.New()

func HashString(input string) string {
	hashInst.Reset()
	hashInst.Write([]byte(input))
	bs := hashInst.Sum(nil)
	return base64.StdEncoding.EncodeToString(bs)
}

func GetMapValueStmt(m map[any]any) *MapValueStmt {
	stmt := &MapValueStmt{}

	for key, val := range m {
		stmt.add(&MapEntryStmt{
			Key:   &PrimitiveValueStmt{RawValue: key, Primitive: GetValuePrimitive(key)},
			Value: &PrimitiveValueStmt{RawValue: val, Primitive: GetValuePrimitive(val)},
		})
	}

	return stmt
}

func GetValuePrimitive(v interface{}) Primitive {
	switch v.(type) {
	case string:
		return Primitive_String
	case uint8:
		return Primitive_Uint8
	case uint16:
		return Primitive_Uint16
	case uint32:
		return Primitive_Uint32
	case uint64:
		return Primitive_Uint64
	case int8:
		return Primitive_Int8
	case int16:
		return Primitive_Int16
	case int32:
		return Primitive_Int32
	case int64:
		return Primitive_Int64
	case float32:
		return Primitive_Float32
	case float64:
		return Primitive_Float64
	case bool:
		return Primitive_Bool
	default:
		panic(fmt.Sprintf("unable to get primitive of %v", v))
	}
}

func GetField(index int, name string, valueType string, nullable bool, metadata map[any]any, defaultValue any) *FieldStmt {
	stmt := &FieldStmt{
		Index: &PrimitiveValueStmt{RawValue: int64(index), Primitive: Primitive_Int64},
		Name:  &IdentifierStmt{Lit: name},
		ValueType: &ValueTypeStmt{
			Ident:    &IdentifierStmt{Lit: valueType},
			Nullable: nullable,
		},
	}

	if metadata != nil {
		stmt.Metadata = GetMapValueStmt(metadata)
	}

	if defaultValue != nil {
		stmt.DefaultValue = &PrimitiveValueStmt{RawValue: defaultValue, Primitive: GetValuePrimitive(defaultValue)}
	}

	return stmt
}

func String(s string) *string {
	return &s
}
