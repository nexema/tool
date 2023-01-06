package internal

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseImportGroup(t *testing.T) {
	var tests = []struct {
		input  string
		expect *[]*ImportStmt
		err    error
	}{
		{
			input: `import: "hello"`,
			expect: &[]*ImportStmt{
				{
					path: &IdentifierStmt{lit: "hello"},
				},
			},
			err: nil,
		},
		{
			input: `import: "my/path" as my_alias`,
			expect: &[]*ImportStmt{
				{
					path:  &IdentifierStmt{lit: "my/path"},
					alias: &IdentifierStmt{lit: "my_alias"},
				},
			},
			err: nil,
		},
		{
			input:  `import: "my/path" as 1231`,
			expect: new([]*ImportStmt),
			err:    errors.New(`1:21 -> expected "ident", given "int" (1231)`),
		},
		{
			input: `import: 
						"my/path" as path
						"second"
						"my_path/another" as another`,
			expect: &[]*ImportStmt{
				{
					path:  &IdentifierStmt{lit: "my/path"},
					alias: &IdentifierStmt{lit: "path"},
				},
				{
					path: &IdentifierStmt{lit: "second"},
				},
				{
					path:  &IdentifierStmt{lit: "my_path/another"},
					alias: &IdentifierStmt{lit: "another"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))

			err := parser.parseImportGroup()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, parser.imports)
		})
	}
}

func TestParseIdentifier(t *testing.T) {
	var tests = []struct {
		input  string
		expect *IdentifierStmt
		err    error
	}{
		{
			input:  `string`,
			expect: &IdentifierStmt{lit: "string"},
			err:    nil,
		},
		{
			input:  `true`,
			expect: &IdentifierStmt{lit: "true"},
			err:    nil,
		},
		{
			input:  `my_path.My_Enum`,
			expect: &IdentifierStmt{alias: "my_path", lit: "My_Enum"},
			err:    nil,
		},
		{
			input:  `my_path.My_Enum.value`,
			expect: &IdentifierStmt{alias: "my_path", lit: "My_Enum"},
			err:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			parser.next()

			ident, err := parser.parseIdentifier()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ident)
		})
	}
}

func TestParseList(t *testing.T) {
	var tests = []struct {
		input  string
		expect *ListStmt
		err    error
	}{
		{
			input: `["my string", true, false, null, 128, 12.4]`,
			expect: &ListStmt{
				&PrimitiveValueStmt{value: "my string", kind: Primitive_String},
				&PrimitiveValueStmt{value: true, kind: Primitive_Bool},
				&PrimitiveValueStmt{value: false, kind: Primitive_Bool},
				&PrimitiveValueStmt{value: nil, kind: Primitive_Null},
				&PrimitiveValueStmt{value: int64(128), kind: Primitive_Int64},
				&PrimitiveValueStmt{value: float64(12.4), kind: Primitive_Float64},
			},
			err: nil,
		},
		{
			input: `"my string", true, false, null, 128, 12.4]`,
			err:   errors.New(`1:0 -> expected "[", given "string" (my string)`),
		},
		{
			input: `["my string", true, MyEnum.unknown, null, 128, 12.4]`,
			expect: &ListStmt{
				&PrimitiveValueStmt{value: "my string", kind: Primitive_String},
				&PrimitiveValueStmt{value: true, kind: Primitive_Bool},
				&TypeValueStmt{value: &IdentifierStmt{lit: "unknown"}, typeName: &IdentifierStmt{lit: "MyEnum"}},
				&PrimitiveValueStmt{value: nil, kind: Primitive_Null},
				&PrimitiveValueStmt{value: int64(128), kind: Primitive_Int64},
				&PrimitiveValueStmt{value: float64(12.4), kind: Primitive_Float64},
			},
		},
		{
			input: `["my string", true,]`,
			err:   errors.New(`1:19 -> expected string, int, float or identifier, given "]"`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			parser.next()

			list, err := parser.parseList()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, list)
		})
	}
}

func TestParseMap(t *testing.T) {
	var tests = []struct {
		input  string
		expect *MapStmt
		err    error
	}{
		{
			input: `[("string":22.43),(true: 23),(13.23: "hello world"),(13: null)]`,
			expect: &MapStmt{
				{
					key:   &PrimitiveValueStmt{value: "string", kind: Primitive_String},
					value: &PrimitiveValueStmt{value: float64(22.43), kind: Primitive_Float64},
				},
				{
					key:   &PrimitiveValueStmt{value: true, kind: Primitive_Bool},
					value: &PrimitiveValueStmt{value: int64(23), kind: Primitive_Int64},
				},
				{
					key:   &PrimitiveValueStmt{value: float64(13.23), kind: Primitive_Float64},
					value: &PrimitiveValueStmt{value: "hello world", kind: Primitive_String},
				},
				{
					key:   &PrimitiveValueStmt{value: int64(13), kind: Primitive_Int64},
					value: &PrimitiveValueStmt{value: nil, kind: Primitive_Null},
				},
			},
			err: nil,
		},
		{
			input: `("string":22.43),(true: 23),(13.23: "hello world"),(13: null)]`,
			err:   errors.New(`1:0 -> expected "[", given "(" (()`),
		},
		{
			input: `["string":22.43),(true: 23),(13.23: "hello world"),(13: null)]`,
			err:   errors.New(`1:1 -> expected "(", given "string" (string)`),
		},
		{
			input: `[("string":22.43)(true: 23),(13.23: "hello world"),(13: null)]`,
			err:   errors.New(`1:17 -> expected "]", given "(" (()`),
		},
		{
			input: `[("string"22.43),(true: 23),(13.23: "hello world"),(13: null)]`,
			err:   errors.New(`1:10 -> expected ":", given "float" (22.43)`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			parser.next()

			m, err := parser.parseMap()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, m)
		})
	}
}

func TestParseValue(t *testing.T) {
	var tests = []struct {
		input  string
		expect ValueStmt
		err    error
	}{
		{
			input:  `"hello world"`,
			expect: &PrimitiveValueStmt{value: "hello world", kind: Primitive_String},
			err:    nil,
		},
		{
			input:  `17.12`,
			expect: &PrimitiveValueStmt{value: float64(17.12), kind: Primitive_Float64},
			err:    nil,
		},
		{
			input:  `17`,
			expect: &PrimitiveValueStmt{value: int64(17), kind: Primitive_Int64},
			err:    nil,
		},
		{
			input:  `true`,
			expect: &PrimitiveValueStmt{value: true, kind: Primitive_Bool},
			err:    nil,
		},
		{
			input:  `false`,
			expect: &PrimitiveValueStmt{value: false, kind: Primitive_Bool},
			err:    nil,
		},
		{
			input:  `null`,
			expect: &PrimitiveValueStmt{value: nil, kind: Primitive_Null},
			err:    nil,
		},
		{
			input:  `MyEnum.unknown`,
			expect: &TypeValueStmt{typeName: &IdentifierStmt{lit: "MyEnum"}, value: &IdentifierStmt{lit: "unknown"}},
			err:    nil,
		},
		{
			input:  `file.MyEnum.unknown`,
			expect: &TypeValueStmt{typeName: &IdentifierStmt{alias: "file", lit: "MyEnum"}, value: &IdentifierStmt{lit: "unknown"}},
			err:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			parser.next()

			ident, err := parser.parseValue()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ident)
		})
	}
}
