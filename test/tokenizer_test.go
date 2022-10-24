package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"tomasweigenast.com/schema_interpreter/internal"
)

func TestStructFieldTokenizer(t *testing.T) {
	tokenizer := internal.NewStructFieldTokenizer()
	type test struct {
		input string
		want  *internal.StructFieldTokenizerResult
	}

	tests := []test{
		{
			input: "b:string 1",
			want: &internal.StructFieldTokenizerResult{
				FieldName:              "b",
				PrimitiveFieldTypeName: "string",
				FieldIndex:             "1",
			},
		},
		{
			input: `b:string 2 = "a default value"`,
			want:  &internal.StructFieldTokenizerResult{FieldName: "b", PrimitiveFieldTypeName: "string"},
		},
		{
			input: `a:boolean 0 = true @[("one":true), ("two":2), ("three":"str")]`,
			want: &internal.StructFieldTokenizerResult{
				FieldName:              "a",
				PrimitiveFieldTypeName: "boolean",
				FieldIndex:             "0",
				DefaultValue:           "true",
				Metadata:               `[("one":true), ("two":2), ("three":"str")]`,
			},
		},
		{
			input: `a:list(int32) 0 = [15, 36,25, 12]`,
			want: &internal.StructFieldTokenizerResult{
				FieldName:              "a",
				PrimitiveFieldTypeName: "list",
				TypeArguments:          []string{"int32"},
				FieldIndex:             "0",
				DefaultValue:           "[15, 36,25, 12]",
			},
		},
		{
			input: `a:map(string,int32) 0 = [("one":21), ("two":2), ("three":3)]`,
			want: &internal.StructFieldTokenizerResult{
				FieldName:              "a",
				PrimitiveFieldTypeName: "map",
				TypeArguments:          []string{"string", "int32"},
				FieldIndex:             "0",
				DefaultValue:           `[("one":21), ("two":2), ("three":3)]`,
			},
		},
		{
			input: `a:map(string,int32) 0 = [("one":21), ("two":2), ("three":3)] @[("one":true), ("two":2), ("three":"str")]`,
			want: &internal.StructFieldTokenizerResult{
				FieldName:              "a",
				PrimitiveFieldTypeName: "map",
				TypeArguments:          []string{"string", "int32"},
				FieldIndex:             "0",
				DefaultValue:           `[("one":21), ("two":2), ("three":3)]`,
				Metadata:               `[("one":true), ("two":2), ("three":"str")]`,
			},
		},
	}

	for _, tc := range tests {
		got, err := tokenizer.Tokenize(tc.input)
		assert.Nil(t, err, fmt.Sprintf("input: %v", tc.input))

		if err != nil {
			assert.Equal(t, got, tc.want)
		}
	}
}
