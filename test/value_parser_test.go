package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"tomasweigenast.com/schema_interpreter/internal"
)

func TestValueParser(t *testing.T) {
	valueParser := internal.NewValueParser()

	type test struct {
		input     string
		wantType  internal.FieldTypePrimitive
		wantValue interface{}
	}

	tests := []test{
		{
			input:     `"hello world, its an string 12542.212415 12154 list(nested) but its an string"`,
			wantType:  internal.String,
			wantValue: "hello world, its an string 12542.212415 12154 list(nested) but its an string",
		},
		{
			input:     `125.256`,
			wantType:  internal.Float64,
			wantValue: 125.256,
		},
		{
			input:     `336656121`,
			wantType:  internal.Int64,
			wantValue: 336656121,
		},
		{
			input:     `true`,
			wantType:  internal.Boolean,
			wantValue: true,
		},
		{
			input:     `false`,
			wantType:  internal.Boolean,
			wantValue: false,
		},
		{
			input:     `[12.1, 25, 32, 54]`,
			wantType:  internal.List,
			wantValue: []interface{}{12.1, 25, 32, 54},
		},
		{
			input:    `[("hello":12.321),(25:true),("sen te":-32.2)]`,
			wantType: internal.Map,
			wantValue: map[interface{}]interface{}{
				"hello":  12.321,
				25:       true,
				"sen te": -32.2,
			},
		},
		{
			input:    `[("one": 5), ("two":2), ("three":12)]`,
			wantType: internal.Map,
			wantValue: map[interface{}]interface{}{
				"one":   5,
				"two":   2,
				"three": 12,
			},
		},
	}

	for _, tc := range tests {
		gotValue, gotType, err := valueParser.ParseString(tc.input)
		assert.Nil(t, err)

		if err == nil {
			assert.Equal(t, gotValue, tc.wantValue)
			assert.Equal(t, gotType, tc.wantType)
		}
	}
}
