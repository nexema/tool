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
		expect *ListValueStmt
		err    error
	}{
		{
			input: `["my string", true, false, null, 128, 12.4]`,
			expect: &ListValueStmt{
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
			expect: &ListValueStmt{
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
		expect *MapValueStmt
		err    error
	}{
		{
			input: `[("string":22.43),(true: 23),(13.23: "hello world"),(13: null)]`,
			expect: &MapValueStmt{
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

func TestParseType(t *testing.T) {
	var tests = []struct {
		input  string
		name   string
		expect *TypeStmt
		err    error
		run    bool
	}{
		{
			name: "without modifier",
			input: `
			type MyType {}
			`,
			expect: &TypeStmt{
				name:     &IdentifierStmt{lit: "MyType"},
				modifier: Token_Struct,
			},
			err: nil,
		},
		{
			name: "struct modifier",
			input: `
			type MyType struct {}
			`,
			expect: &TypeStmt{
				name:     &IdentifierStmt{lit: "MyType"},
				modifier: Token_Struct,
			},
			err: nil,
		},
		{
			name: "enum modifier",
			input: `
			type My_Type enum {}
			`,
			expect: &TypeStmt{
				name:     &IdentifierStmt{lit: "My_Type"},
				modifier: Token_Enum,
			},
			err: nil,
		},
		{
			name: "union modifier",
			input: `
			type MyType union {}
			`,
			expect: &TypeStmt{
				name:     &IdentifierStmt{lit: "MyType"},
				modifier: Token_Union,
			},
			err: nil,
		},
		{
			name: "with metadata",
			input: `
			@[("obsolete": true),("alternative":"MyAnotherType")]
			type MyType union {}
			`,
			expect: &TypeStmt{
				name:     &IdentifierStmt{lit: "MyType"},
				modifier: Token_Union,
				metadata: &MapValueStmt{
					{key: &PrimitiveValueStmt{value: "obsolete", kind: Primitive_String}, value: &PrimitiveValueStmt{value: true, kind: Primitive_Bool}},
					{key: &PrimitiveValueStmt{value: "alternative", kind: Primitive_String}, value: &PrimitiveValueStmt{value: "MyAnotherType", kind: Primitive_String}},
				},
			},
			err: nil,
		},
		{
			name: "full struct type",
			input: `
			type MyType struct {
				1 field_1: string
				2 field_2: bool = true
				3 field_3: int32 = 43 @[("obsolete": true)]
				4 field_4: float32 @[("a": "b")]
				field_5: list(bool?) = [true]
			}
			`,
			expect: &TypeStmt{
				name:     &IdentifierStmt{lit: "MyType"},
				modifier: Token_Struct,
				fields: &[]*FieldStmt{
					{
						index:     &PrimitiveValueStmt{value: int64(1), kind: Primitive_Int64},
						name:      &IdentifierStmt{lit: "field_1"},
						valueType: &ValueTypeStmt{ident: &IdentifierStmt{lit: "string"}},
					},
					{
						index:        &PrimitiveValueStmt{value: int64(2), kind: Primitive_Int64},
						name:         &IdentifierStmt{lit: "field_2"},
						valueType:    &ValueTypeStmt{ident: &IdentifierStmt{lit: "bool"}},
						defaultValue: &PrimitiveValueStmt{value: true, kind: Primitive_Bool},
					},
					{
						index:        &PrimitiveValueStmt{value: int64(3), kind: Primitive_Int64},
						name:         &IdentifierStmt{lit: "field_3"},
						valueType:    &ValueTypeStmt{ident: &IdentifierStmt{lit: "int32"}},
						defaultValue: &PrimitiveValueStmt{value: int64(43), kind: Primitive_Int64},
						metadata: &MapValueStmt{
							{key: &PrimitiveValueStmt{value: "obsolete", kind: Primitive_String}, value: &PrimitiveValueStmt{value: true, kind: Primitive_Bool}},
						},
					},
					{
						index:     &PrimitiveValueStmt{value: int64(4), kind: Primitive_Int64},
						name:      &IdentifierStmt{lit: "field_4"},
						valueType: &ValueTypeStmt{ident: &IdentifierStmt{lit: "float32"}},
						metadata: &MapValueStmt{
							{key: &PrimitiveValueStmt{value: "a", kind: Primitive_String}, value: &PrimitiveValueStmt{value: "b", kind: Primitive_String}},
						},
					},
					{
						name: &IdentifierStmt{lit: "field_5"},
						valueType: &ValueTypeStmt{
							ident: &IdentifierStmt{lit: "list"},
							typeArguments: &[]*ValueTypeStmt{
								{
									ident:    &IdentifierStmt{lit: "bool"},
									nullable: true,
								},
							},
						},
						defaultValue: &ListValueStmt{
							&PrimitiveValueStmt{value: true, kind: Primitive_Bool},
						},
					},
				},
			},
			run: true,
			err: nil,
		},
	}
	for _, tt := range tests {
		// if !tt.run {
		// 	continue
		// }
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			parser.next()

			stmt, err := parser.parseType()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, stmt)
		})
	}
}

func TestParseField(t *testing.T) {
	var tests = []struct {
		input  string
		expect *FieldStmt
		err    error
	}{
		{
			input: `1 field_name: string`,
			expect: &FieldStmt{
				index:     &PrimitiveValueStmt{value: int64(1), kind: Primitive_Int64},
				name:      &IdentifierStmt{lit: "field_name"},
				valueType: &ValueTypeStmt{ident: &IdentifierStmt{lit: "string"}},
			},
			err: nil,
		},
		{
			input: `field_name: string`,
			expect: &FieldStmt{
				name:      &IdentifierStmt{lit: "field_name"},
				valueType: &ValueTypeStmt{ident: &IdentifierStmt{lit: "string"}},
			},
			err: nil,
		},
		{
			input: `field_name: string?`,
			expect: &FieldStmt{
				name:      &IdentifierStmt{lit: "field_name"},
				valueType: &ValueTypeStmt{ident: &IdentifierStmt{lit: "string"}, nullable: true},
			},
			err: nil,
		},
		{
			input: `field_name: string? = null`,
			expect: &FieldStmt{
				name:         &IdentifierStmt{lit: "field_name"},
				valueType:    &ValueTypeStmt{ident: &IdentifierStmt{lit: "string"}, nullable: true},
				defaultValue: &PrimitiveValueStmt{value: nil, kind: Primitive_Null},
			},
			err: nil,
		},
		{
			input: `field_name: string? = "hello world"`,
			expect: &FieldStmt{
				name:         &IdentifierStmt{lit: "field_name"},
				valueType:    &ValueTypeStmt{ident: &IdentifierStmt{lit: "string"}, nullable: true},
				defaultValue: &PrimitiveValueStmt{value: "hello world", kind: Primitive_String},
			},
			err: nil,
		},
		{
			input: `field_name: binary? @[("obsolete":true)]`,
			expect: &FieldStmt{
				name:      &IdentifierStmt{lit: "field_name"},
				valueType: &ValueTypeStmt{ident: &IdentifierStmt{lit: "binary"}, nullable: true},
				metadata: &MapValueStmt{
					{key: &PrimitiveValueStmt{value: "obsolete", kind: Primitive_String}, value: &PrimitiveValueStmt{value: true, kind: Primitive_Bool}},
				},
			},
			err: nil,
		},
		{
			input: `5 field_name: int = 54 @[("obsolete":true)]`,
			expect: &FieldStmt{
				index:     &PrimitiveValueStmt{value: int64(5), kind: Primitive_Int64},
				name:      &IdentifierStmt{lit: "field_name"},
				valueType: &ValueTypeStmt{ident: &IdentifierStmt{lit: "int"}},
				metadata: &MapValueStmt{
					{key: &PrimitiveValueStmt{value: "obsolete", kind: Primitive_String}, value: &PrimitiveValueStmt{value: true, kind: Primitive_Bool}},
				},
				defaultValue: &PrimitiveValueStmt{value: int64(54), kind: Primitive_Int64},
			},
			err: nil,
		},
		{
			input: `5 field_name: list(string?) = ["hello", null] @[("obsolete":true)]`,
			expect: &FieldStmt{
				index: &PrimitiveValueStmt{value: int64(5), kind: Primitive_Int64},
				name:  &IdentifierStmt{lit: "field_name"},
				valueType: &ValueTypeStmt{ident: &IdentifierStmt{lit: "list"}, typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{lit: "string"}, nullable: true},
				}},
				metadata: &MapValueStmt{
					{key: &PrimitiveValueStmt{value: "obsolete", kind: Primitive_String}, value: &PrimitiveValueStmt{value: true, kind: Primitive_Bool}},
				},
				defaultValue: &ListValueStmt{
					&PrimitiveValueStmt{value: "hello", kind: Primitive_String},
					&PrimitiveValueStmt{value: nil, kind: Primitive_Null},
				},
			},
			err: nil,
		},
		{
			input: `5 field_name: map(string, bool?) = [("hello": null), ("second": true)] @[("obsolete":true)]`,
			expect: &FieldStmt{
				index: &PrimitiveValueStmt{value: int64(5), kind: Primitive_Int64},
				name:  &IdentifierStmt{lit: "field_name"},
				valueType: &ValueTypeStmt{ident: &IdentifierStmt{lit: "map"}, typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{lit: "string"}},
					{ident: &IdentifierStmt{lit: "bool"}, nullable: true},
				}},
				metadata: &MapValueStmt{
					{key: &PrimitiveValueStmt{value: "obsolete", kind: Primitive_String}, value: &PrimitiveValueStmt{value: true, kind: Primitive_Bool}},
				},
				defaultValue: &MapValueStmt{
					{
						key:   &PrimitiveValueStmt{kind: Primitive_String, value: "hello"},
						value: &PrimitiveValueStmt{kind: Primitive_Null, value: nil},
					},
					{
						key:   &PrimitiveValueStmt{kind: Primitive_String, value: "second"},
						value: &PrimitiveValueStmt{kind: Primitive_Bool, value: true},
					},
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			parser.next()

			stmt, err := parser.parseField()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, stmt)
		})
	}
}

func TestParseValueType(t *testing.T) {
	var tests = []struct {
		input  string
		expect *ValueTypeStmt
		err    error
	}{
		{
			input:  `string`,
			expect: &ValueTypeStmt{ident: &IdentifierStmt{lit: "string"}},
			err:    nil,
		},
		{
			input:  `string?`,
			expect: &ValueTypeStmt{ident: &IdentifierStmt{lit: "string"}, nullable: true},
			err:    nil,
		},
		{
			input: `list(string)`,
			expect: &ValueTypeStmt{
				ident: &IdentifierStmt{lit: "list"},
				typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{lit: "string"}},
				},
			},
			err: nil,
		},
		{
			input: `list(string,bool)`,
			expect: &ValueTypeStmt{
				ident: &IdentifierStmt{lit: "list"},
				typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{lit: "string"}},
					{ident: &IdentifierStmt{lit: "bool"}},
				},
			},
			err: nil,
		},
		{
			input: `list(string?)`,
			expect: &ValueTypeStmt{
				ident: &IdentifierStmt{lit: "list"},
				typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{lit: "string"}, nullable: true},
				},
			},
			err: nil,
		},
		{
			input: `list(string?, bool?)`,
			expect: &ValueTypeStmt{
				ident: &IdentifierStmt{lit: "list"},
				typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{lit: "string"}, nullable: true},
					{ident: &IdentifierStmt{lit: "bool"}, nullable: true},
				},
			},
			err: nil,
		},
		{
			input: `map(int64, bool?)`,
			expect: &ValueTypeStmt{
				ident: &IdentifierStmt{lit: "map"},
				typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{lit: "int64"}, nullable: false},
					{ident: &IdentifierStmt{lit: "bool"}, nullable: true},
				},
			},
			err: nil,
		},
		{
			input: `list(int64)?`,
			expect: &ValueTypeStmt{
				ident: &IdentifierStmt{lit: "list"},
				typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{lit: "int64"}, nullable: false},
				},
				nullable: true,
			},
			err: nil,
		},
		{
			input: `list(int64?)?`,
			expect: &ValueTypeStmt{
				ident: &IdentifierStmt{lit: "list"},
				typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{lit: "int64"}, nullable: true},
				},
				nullable: true,
			},
			err: nil,
		},
		{
			input: `list(MyEnum)?`,
			expect: &ValueTypeStmt{
				ident: &IdentifierStmt{lit: "list"},
				typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{lit: "MyEnum"}},
				},
				nullable: true,
			},
			err: nil,
		},
		{
			input: `list(another.MyEnum)?`,
			expect: &ValueTypeStmt{
				ident: &IdentifierStmt{lit: "list"},
				typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{alias: "another", lit: "MyEnum"}},
				},
				nullable: true,
			},
			err: nil,
		},
		{
			input: `map(string, another.MyEnum?)?`,
			expect: &ValueTypeStmt{
				ident: &IdentifierStmt{lit: "map"},
				typeArguments: &[]*ValueTypeStmt{
					{ident: &IdentifierStmt{lit: "string"}},
					{ident: &IdentifierStmt{alias: "another", lit: "MyEnum"}, nullable: true},
				},
				nullable: true,
			},
			err: nil,
		},
		{
			input: `map(string, another.MyEnum??`,
			err:   errors.New(`1:27 -> expected ")", given "?" (?)`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			parser.next()

			stmt, err := parser.parseValueTypeStmt()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, stmt)
		})
	}
}

func TestParse(t *testing.T) {
	var tests = []struct {
		name   string
		input  string
		expect *Ast
		err    error
	}{
		{
			name: "parse imports",
			input: `
			import:
				"my_file.nex"
				"another.nex" as another`,
			expect: &Ast{
				imports: &[]*ImportStmt{
					{path: &IdentifierStmt{lit: "my_file.nex"}},
					{path: &IdentifierStmt{lit: "another.nex"}, alias: &IdentifierStmt{lit: "another"}},
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.Parse()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ast)
		})
	}
}
