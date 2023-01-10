package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateType(t *testing.T) {
	var tests = []struct {
		name   string
		input  *TypeStmt
		errors *ErrorCollection
	}{
		{
			name: "rule 2 success on struct or union if fields index are unique",
			input: &TypeStmt{
				name: &IdentifierStmt{lit: "A"},
				fields: &[]*FieldStmt{
					{index: &PrimitiveValueStmt{value: int64(1), kind: Primitive_Int64}},
					{index: &PrimitiveValueStmt{value: int64(2), kind: Primitive_Int64}},
					{index: &PrimitiveValueStmt{value: int64(3), kind: Primitive_Int64}},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 2 fails on struct or union if fields index are not unique",
			input: &TypeStmt{
				name: &IdentifierStmt{lit: "A"},
				fields: &[]*FieldStmt{
					{index: &PrimitiveValueStmt{value: int64(1), kind: Primitive_Int64}},
					{index: &PrimitiveValueStmt{value: int64(2), kind: Primitive_Int64}},
					{index: &PrimitiveValueStmt{value: int64(1), kind: Primitive_Int64}},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> index 1 already defined for a field"),
			},
		},
		{
			name: "rule 2 fails on enum if fields index does not start with 0",
			input: &TypeStmt{
				name:     &IdentifierStmt{lit: "A"},
				modifier: Token_Enum,
				fields: &[]*FieldStmt{
					{index: &PrimitiveValueStmt{value: int64(1), kind: Primitive_Int64}},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> expected the first field in an enum to have the index 0, given index 1"),
			},
		},
		{
			name: "rule 2 fails on enum if fields index are not correlative",
			input: &TypeStmt{
				name:     &IdentifierStmt{lit: "A"},
				modifier: Token_Enum,
				fields: &[]*FieldStmt{
					{index: &PrimitiveValueStmt{value: int64(0), kind: Primitive_Int64}},
					{index: &PrimitiveValueStmt{value: int64(1), kind: Primitive_Int64}},
					{index: &PrimitiveValueStmt{value: int64(3), kind: Primitive_Int64}},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> field indexes in an enum must be correlative"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer([]*ResolvedContext{})
			analyzer.skipFields = true
			analyzer.currentContext = &ResolvedContext{
				owner: &Ast{file: &File{pkg: "root"}},
			}
			analyzer.validateType(tt.input)
			require.Equal(t, tt.errors, analyzer.errors)
		})
	}
}

func TestValidateField(t *testing.T) {
	var tests = []struct {
		name           string
		input          *FieldStmt
		currentContext *ResolvedContext
		errors         *ErrorCollection
	}{
		{
			name: "rule 1.a success with exactly one type argument",
			input: &FieldStmt{
				valueType: &ValueTypeStmt{
					ident: &IdentifierStmt{lit: "list"},
					typeArguments: &[]*ValueTypeStmt{
						{ident: &IdentifierStmt{lit: "string"}},
					},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 1.a fails if zero type argument is specified",
			input: &FieldStmt{
				valueType: &ValueTypeStmt{
					ident:         &IdentifierStmt{lit: "list"},
					typeArguments: &[]*ValueTypeStmt{},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> list expects exactly one type argument"),
			},
		},
		{
			name: "rule 1.a fails with more than one type argument",
			input: &FieldStmt{
				valueType: &ValueTypeStmt{
					ident: &IdentifierStmt{lit: "list"},
					typeArguments: &[]*ValueTypeStmt{
						{ident: &IdentifierStmt{lit: "string"}},
						{ident: &IdentifierStmt{lit: "int64"}},
					},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> list expects exactly one type argument"),
			},
		},
		{
			name: "rule 1.b success with exactly two type arguments",
			input: &FieldStmt{
				valueType: &ValueTypeStmt{
					ident: &IdentifierStmt{lit: "map"},
					typeArguments: &[]*ValueTypeStmt{
						{ident: &IdentifierStmt{lit: "string"}},
						{ident: &IdentifierStmt{lit: "int64"}},
					},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 1.b fails with more than two type arguments",
			input: &FieldStmt{
				valueType: &ValueTypeStmt{
					ident: &IdentifierStmt{lit: "map"},
					typeArguments: &[]*ValueTypeStmt{
						{ident: &IdentifierStmt{lit: "string"}},
						{ident: &IdentifierStmt{lit: "int64"}},
						{ident: &IdentifierStmt{lit: "bool"}},
					},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> map expects exactly two type arguments"),
			},
		},
		{
			name: "rule 1.b fails with zero type arguments",
			input: &FieldStmt{
				valueType: &ValueTypeStmt{
					ident:         &IdentifierStmt{lit: "map"},
					typeArguments: &[]*ValueTypeStmt{},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> map expects exactly two type arguments"),
			},
		},
		{
			name: "rule 2 fails if default value does not match value type",
			input: &FieldStmt{
				valueType: &ValueTypeStmt{ident: &IdentifierStmt{lit: "string"}},
				defaultValue: &PrimitiveValueStmt{
					value: int64(25),
					kind:  Primitive_Int64,
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> field's default value is not of type string, it is int64"),
			},
		},
		{
			name: "rule 2 success if default value matches value type",
			input: &FieldStmt{
				valueType: &ValueTypeStmt{ident: &IdentifierStmt{lit: "int64"}},
				defaultValue: &PrimitiveValueStmt{
					value: int64(25),
					kind:  Primitive_Int64,
				},
			},
			errors: &ErrorCollection{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer([]*ResolvedContext{})
			analyzer.validateField(tt.input)

			if tt.currentContext != nil {
				analyzer.currentContext = tt.currentContext
			}

			require.Equal(t, tt.errors, analyzer.errors)
		})
	}
}

func TestValidateMetadata(t *testing.T) {
	var tests = []struct {
		name   string
		input  *MapValueStmt
		errors *ErrorCollection
	}{
		{
			name: "rule 1 success",
			input: &MapValueStmt{
				{
					key:   &PrimitiveValueStmt{value: "string", kind: Primitive_String},
					value: &PrimitiveValueStmt{value: "string", kind: Primitive_String},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 1 fails if not string",
			input: &MapValueStmt{
				{
					key: &PrimitiveValueStmt{kind: Primitive_Int64},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> metadata map keys must be of type string"),
			},
		},
		{
			name: "rule 2 success if value is string, bool, int64 or float64",
			input: &MapValueStmt{
				{
					key:   &PrimitiveValueStmt{value: "1", kind: Primitive_String},
					value: &PrimitiveValueStmt{kind: Primitive_String},
				},
				{
					key:   &PrimitiveValueStmt{value: "2", kind: Primitive_String},
					value: &PrimitiveValueStmt{kind: Primitive_Bool},
				},
				{
					key:   &PrimitiveValueStmt{value: "3", kind: Primitive_String},
					value: &PrimitiveValueStmt{kind: Primitive_Int64},
				},
				{
					key:   &PrimitiveValueStmt{value: "4", kind: Primitive_String},
					value: &PrimitiveValueStmt{kind: Primitive_Float64},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 2 fails if value is not string, bool, int64 or float64",
			input: &MapValueStmt{
				{
					key:   &PrimitiveValueStmt{value: "1", kind: Primitive_String},
					value: &PrimitiveValueStmt{kind: Primitive_List},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> metadata map values must be one of the following types: string|bool|float64|int64, given: list"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer([]*ResolvedContext{})
			analyzer.validateMetadata(tt.input)
			require.Equal(t, tt.errors, analyzer.errors)
		})
	}
}

func TestValidateMap(t *testing.T) {
	var tests = []struct {
		name   string
		input  *MapValueStmt
		errors *ErrorCollection
	}{
		{
			name: "rule 1 success",
			input: &MapValueStmt{
				{
					key:   &PrimitiveValueStmt{value: "string", kind: Primitive_String},
					value: &PrimitiveValueStmt{value: "string", kind: Primitive_String},
				},
			},
			errors: &ErrorCollection{},
		},
		{
			name: "rule 1 fails with list key",
			input: &MapValueStmt{
				{
					key: &ListValueStmt{
						&PrimitiveValueStmt{kind: Primitive_String},
					},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> map's keys cannot be of type list, map, null or a custom type"),
			},
		},
		{
			name: "rule 1 fails with map key",
			input: &MapValueStmt{
				{
					key: &MapValueStmt{
						{
							key: &PrimitiveValueStmt{kind: Primitive_String},
						},
					},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> map's keys cannot be of type list, map, null or a custom type"),
			},
		},
		{
			name: "rule 1 fails with null key",
			input: &MapValueStmt{
				{
					key: &PrimitiveValueStmt{kind: Primitive_Null},
				},
			},
			errors: &ErrorCollection{
				errors.New("[analyzer] 0:0 -> map's keys cannot be of type list, map, null or a custom type"),
			},
		},
		{
			name: "rule 2 with string key",
			input: &MapValueStmt{
				{
					key: &PrimitiveValueStmt{value: "a", kind: Primitive_String},
				},
				{
					key: &PrimitiveValueStmt{value: "a", kind: Primitive_String},
				},
			},
			errors: &ErrorCollection{
				errors.New(`[analyzer] 0:0 -> key "a" already exists in map`),
			},
		},
		{
			name: "rule 2 with int key",
			input: &MapValueStmt{
				{
					key: &PrimitiveValueStmt{value: int64(2), kind: Primitive_Int64},
				},
				{
					key: &PrimitiveValueStmt{value: int64(2), kind: Primitive_Int64},
				},
			},
			errors: &ErrorCollection{
				errors.New(`[analyzer] 0:0 -> key "2" already exists in map`),
			},
		},
		{
			name: "rule 2 with float key",
			input: &MapValueStmt{
				{
					key: &PrimitiveValueStmt{value: float64(2.5), kind: Primitive_Float64},
				},
				{
					key: &PrimitiveValueStmt{value: float64(2.5), kind: Primitive_Float64},
				},
			},
			errors: &ErrorCollection{
				errors.New(`[analyzer] 0:0 -> key "2.5" already exists in map`),
			},
		},
		{
			name: "rule 2 with bool key",
			input: &MapValueStmt{
				{
					key: &PrimitiveValueStmt{value: true, kind: Primitive_Bool},
				},
				{
					key: &PrimitiveValueStmt{value: true, kind: Primitive_Bool},
				},
			},
			errors: &ErrorCollection{
				errors.New(`[analyzer] 0:0 -> key "true" already exists in map`),
			},
		},
		{
			name: "rule 2 with enum key",
			input: &MapValueStmt{
				{
					key: &TypeValueStmt{
						typeName: &IdentifierStmt{lit: "MyEnum"},
						value:    &IdentifierStmt{lit: "red"},
					},
				},
				{
					key: &TypeValueStmt{
						typeName: &IdentifierStmt{lit: "MyEnum"},
						value:    &IdentifierStmt{lit: "red"},
					},
				},
			},
			errors: &ErrorCollection{
				errors.New(`[analyzer] 0:0 -> key "MyEnum.red" already exists in map`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer([]*ResolvedContext{})
			analyzer.validateMap(tt.input)
			require.Equal(t, tt.errors, analyzer.errors)
		})
	}
}
