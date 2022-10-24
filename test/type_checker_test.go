package test

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"tomasweigenast.com/schema_interpreter/internal"
)

func TestTypeChecker(t *testing.T) {
	typeChecker := internal.NewTypeChecker()

	type test struct {
		input interface{}
		want  internal.FieldTypePrimitive
	}

	tests := []test{
		{
			input: "hello world",
			want:  internal.String,
		},
		{
			input: true,
			want:  internal.Boolean,
		},
		{
			input: false,
			want:  internal.Boolean,
		},
		{
			input: int64(math.MinInt8),
			want:  internal.Int8,
		},
		{
			input: int64(math.MaxUint8),
			want:  internal.Uint8,
		},
		{
			input: uint64(math.MaxUint8),
			want:  internal.Uint8,
		},
		{
			input: int64(math.MaxInt16),
			want:  internal.Int16,
		},
		{
			input: int64(math.MaxUint16),
			want:  internal.Uint16,
		},
		{
			input: uint64(math.MaxUint16),
			want:  internal.Uint16,
		},
		{
			input: int64(math.MaxInt32),
			want:  internal.Int32,
		},
		{
			input: int64(math.MaxUint32),
			want:  internal.Uint32,
		},
		{
			input: uint64(math.MaxUint32),
			want:  internal.Uint32,
		},
		{
			input: int64(math.MaxInt64),
			want:  internal.Int64,
		},
		{
			input: uint64(math.MaxUint64),
			want:  internal.Uint64,
		},
		{
			input: math.MaxFloat64,
			want:  internal.Float64,
		},
		{
			input: math.MaxFloat32,
			want:  internal.Float32,
		},
	}

	for i, tc := range tests {
		_, err := typeChecker.ValidateType(tc.input, tc.want)
		assert.Nil(t, err, "test: "+fmt.Sprint(i+1))
		// assert.Equal(t, got, tc.wantValue)
	}
}
