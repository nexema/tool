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
					Path: &IdentifierStmt{Lit: "hello"},
				},
			},
			err: nil,
		},
		{
			input: `import: "my/path" as my_alias`,
			expect: &[]*ImportStmt{
				{
					Path:  &IdentifierStmt{Lit: "my/path"},
					Alias: &IdentifierStmt{Lit: "my_alias"},
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
					Path:  &IdentifierStmt{Lit: "my/path"},
					Alias: &IdentifierStmt{Lit: "path"},
				},
				{
					Path: &IdentifierStmt{Lit: "second"},
				},
				{
					Path:  &IdentifierStmt{Lit: "my_path/another"},
					Alias: &IdentifierStmt{Lit: "another"},
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
			expect: &IdentifierStmt{Lit: "string"},
			err:    nil,
		},
		{
			input:  `true`,
			expect: &IdentifierStmt{Lit: "true"},
			err:    nil,
		},
		{
			input:  `my_path.My_Enum`,
			expect: &IdentifierStmt{Alias: "my_path", Lit: "My_Enum"},
			err:    nil,
		},
		{
			input:  `my_path.My_Enum.value`,
			expect: &IdentifierStmt{Alias: "my_path", Lit: "My_Enum"},
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
				&PrimitiveValueStmt{RawValue: "my string", Primitive: Primitive_String},
				&PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool},
				&PrimitiveValueStmt{RawValue: false, Primitive: Primitive_Bool},
				&PrimitiveValueStmt{RawValue: nil, Primitive: Primitive_Null},
				&PrimitiveValueStmt{RawValue: int64(128), Primitive: Primitive_Int64},
				&PrimitiveValueStmt{RawValue: float64(12.4), Primitive: Primitive_Float64},
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
				&PrimitiveValueStmt{RawValue: "my string", Primitive: Primitive_String},
				&PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool},
				&TypeValueStmt{RawValue: &IdentifierStmt{Lit: "unknown"}, TypeName: &IdentifierStmt{Lit: "MyEnum"}},
				&PrimitiveValueStmt{RawValue: nil, Primitive: Primitive_Null},
				&PrimitiveValueStmt{RawValue: int64(128), Primitive: Primitive_Int64},
				&PrimitiveValueStmt{RawValue: float64(12.4), Primitive: Primitive_Float64},
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
					Key:   &PrimitiveValueStmt{RawValue: "string", Primitive: Primitive_String},
					Value: &PrimitiveValueStmt{RawValue: float64(22.43), Primitive: Primitive_Float64},
				},
				{
					Key:   &PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool},
					Value: &PrimitiveValueStmt{RawValue: int64(23), Primitive: Primitive_Int64},
				},
				{
					Key:   &PrimitiveValueStmt{RawValue: float64(13.23), Primitive: Primitive_Float64},
					Value: &PrimitiveValueStmt{RawValue: "hello world", Primitive: Primitive_String},
				},
				{
					Key:   &PrimitiveValueStmt{RawValue: int64(13), Primitive: Primitive_Int64},
					Value: &PrimitiveValueStmt{RawValue: nil, Primitive: Primitive_Null},
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
			expect: &PrimitiveValueStmt{RawValue: "hello world", Primitive: Primitive_String},
			err:    nil,
		},
		{
			input:  `17.12`,
			expect: &PrimitiveValueStmt{RawValue: float64(17.12), Primitive: Primitive_Float64},
			err:    nil,
		},
		{
			input:  `17`,
			expect: &PrimitiveValueStmt{RawValue: int64(17), Primitive: Primitive_Int64},
			err:    nil,
		},
		{
			input:  `true`,
			expect: &PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool},
			err:    nil,
		},
		{
			input:  `false`,
			expect: &PrimitiveValueStmt{RawValue: false, Primitive: Primitive_Bool},
			err:    nil,
		},
		{
			input:  `null`,
			expect: &PrimitiveValueStmt{RawValue: nil, Primitive: Primitive_Null},
			err:    nil,
		},
		{
			input:  `MyEnum.unknown`,
			expect: &TypeValueStmt{TypeName: &IdentifierStmt{Lit: "MyEnum"}, RawValue: &IdentifierStmt{Lit: "unknown"}},
			err:    nil,
		},
		{
			input:  `file.MyEnum.unknown`,
			expect: &TypeValueStmt{TypeName: &IdentifierStmt{Alias: "file", Lit: "MyEnum"}, RawValue: &IdentifierStmt{Lit: "unknown"}},
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
	}{
		{
			name: "without modifier",
			input: `
			type MyType {}
			`,
			expect: &TypeStmt{
				Name:          &IdentifierStmt{Lit: "MyType"},
				Modifier:      Token_Struct,
				Documentation: new([]*CommentStmt),
			},
			err: nil,
		},
		{
			name: "struct modifier",
			input: `
			type MyType struct {}
			`,
			expect: &TypeStmt{
				Name:          &IdentifierStmt{Lit: "MyType"},
				Modifier:      Token_Struct,
				Documentation: new([]*CommentStmt),
			},
			err: nil,
		},
		{
			name: "enum modifier",
			input: `
			type My_Type enum {}
			`,
			expect: &TypeStmt{
				Name:          &IdentifierStmt{Lit: "My_Type"},
				Modifier:      Token_Enum,
				Documentation: new([]*CommentStmt),
			},
			err: nil,
		},
		{
			name: "union modifier",
			input: `
			type MyType union {}
			`,
			expect: &TypeStmt{
				Name:          &IdentifierStmt{Lit: "MyType"},
				Modifier:      Token_Union,
				Documentation: new([]*CommentStmt),
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
				Name:     &IdentifierStmt{Lit: "MyType"},
				Modifier: Token_Union,
				Metadata: &MapValueStmt{
					{Key: &PrimitiveValueStmt{RawValue: "obsolete", Primitive: Primitive_String}, Value: &PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool}},
					{Key: &PrimitiveValueStmt{RawValue: "alternative", Primitive: Primitive_String}, Value: &PrimitiveValueStmt{RawValue: "MyAnotherType", Primitive: Primitive_String}},
				},
				Documentation: new([]*CommentStmt),
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
				Name:     &IdentifierStmt{Lit: "MyType"},
				Modifier: Token_Struct,
				Fields: &[]*FieldStmt{
					{
						Index:     &PrimitiveValueStmt{RawValue: int64(1), Primitive: Primitive_Int64},
						Name:      &IdentifierStmt{Lit: "field_1"},
						ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "string"}},
					},
					{
						Index:        &PrimitiveValueStmt{RawValue: int64(2), Primitive: Primitive_Int64},
						Name:         &IdentifierStmt{Lit: "field_2"},
						ValueType:    &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "bool"}},
						DefaultValue: &PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool},
					},
					{
						Index:        &PrimitiveValueStmt{RawValue: int64(3), Primitive: Primitive_Int64},
						Name:         &IdentifierStmt{Lit: "field_3"},
						ValueType:    &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "int32"}},
						DefaultValue: &PrimitiveValueStmt{RawValue: int64(43), Primitive: Primitive_Int64},
						Metadata: &MapValueStmt{
							{Key: &PrimitiveValueStmt{RawValue: "obsolete", Primitive: Primitive_String}, Value: &PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool}},
						},
					},
					{
						Index:     &PrimitiveValueStmt{RawValue: int64(4), Primitive: Primitive_Int64},
						Name:      &IdentifierStmt{Lit: "field_4"},
						ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "float32"}},
						Metadata: &MapValueStmt{
							{Key: &PrimitiveValueStmt{RawValue: "a", Primitive: Primitive_String}, Value: &PrimitiveValueStmt{RawValue: "b", Primitive: Primitive_String}},
						},
					},
					{
						Name: &IdentifierStmt{Lit: "field_5"},
						ValueType: &ValueTypeStmt{
							Ident: &IdentifierStmt{Lit: "list"},
							TypeArguments: &[]*ValueTypeStmt{
								{
									Ident:    &IdentifierStmt{Lit: "bool"},
									Nullable: true,
								},
							},
						},
						DefaultValue: &ListValueStmt{
							&PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool},
						},
					},
				},
				Documentation: new([]*CommentStmt),
			},
			err: nil,
		},
		{
			name: "full enum type",
			input: `
			type MyType enum {
				0 unknown
				1 first
				2 second @[("a":23)]
				last @[("a":23)]
			}
			`,
			expect: &TypeStmt{
				Name:     &IdentifierStmt{Lit: "MyType"},
				Modifier: Token_Enum,
				Fields: &[]*FieldStmt{
					{
						Index: &PrimitiveValueStmt{RawValue: int64(0), Primitive: Primitive_Int64},
						Name:  &IdentifierStmt{Lit: "unknown"},
					},
					{
						Index: &PrimitiveValueStmt{RawValue: int64(1), Primitive: Primitive_Int64},
						Name:  &IdentifierStmt{Lit: "first"},
					},
					{
						Index: &PrimitiveValueStmt{RawValue: int64(2), Primitive: Primitive_Int64},
						Name:  &IdentifierStmt{Lit: "second"},
						Metadata: &MapValueStmt{
							{
								Key:   &PrimitiveValueStmt{RawValue: "a", Primitive: Primitive_String},
								Value: &PrimitiveValueStmt{RawValue: int64(23), Primitive: Primitive_Int64},
							},
						},
					},
					{
						Name: &IdentifierStmt{Lit: "last"},
						Metadata: &MapValueStmt{
							{
								Key:   &PrimitiveValueStmt{RawValue: "a", Primitive: Primitive_String},
								Value: &PrimitiveValueStmt{RawValue: int64(23), Primitive: Primitive_Int64},
							},
						},
					},
				},
				Documentation: new([]*CommentStmt),
			},
			err: nil,
		},
	}
	for _, tt := range tests {
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
				Index:     &PrimitiveValueStmt{RawValue: int64(1), Primitive: Primitive_Int64},
				Name:      &IdentifierStmt{Lit: "field_name"},
				ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "string"}},
			},
			err: nil,
		},
		{
			input: `field_name: string`,
			expect: &FieldStmt{
				Name:      &IdentifierStmt{Lit: "field_name"},
				ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "string"}},
			},
			err: nil,
		},
		{
			input: `field_name: string?`,
			expect: &FieldStmt{
				Name:      &IdentifierStmt{Lit: "field_name"},
				ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "string"}, Nullable: true},
			},
			err: nil,
		},
		{
			input: `field_name: string? = null`,
			expect: &FieldStmt{
				Name:         &IdentifierStmt{Lit: "field_name"},
				ValueType:    &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "string"}, Nullable: true},
				DefaultValue: &PrimitiveValueStmt{RawValue: nil, Primitive: Primitive_Null},
			},
			err: nil,
		},
		{
			input: `field_name: string? = "hello world"`,
			expect: &FieldStmt{
				Name:         &IdentifierStmt{Lit: "field_name"},
				ValueType:    &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "string"}, Nullable: true},
				DefaultValue: &PrimitiveValueStmt{RawValue: "hello world", Primitive: Primitive_String},
			},
			err: nil,
		},
		{
			input: `field_name: binary? @[("obsolete":true)]`,
			expect: &FieldStmt{
				Name:      &IdentifierStmt{Lit: "field_name"},
				ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "binary"}, Nullable: true},
				Metadata: &MapValueStmt{
					{Key: &PrimitiveValueStmt{RawValue: "obsolete", Primitive: Primitive_String}, Value: &PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool}},
				},
			},
			err: nil,
		},
		{
			input: `5 field_name: int = 54 @[("obsolete":true)]`,
			expect: &FieldStmt{
				Index:     &PrimitiveValueStmt{RawValue: int64(5), Primitive: Primitive_Int64},
				Name:      &IdentifierStmt{Lit: "field_name"},
				ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "int"}},
				Metadata: &MapValueStmt{
					{Key: &PrimitiveValueStmt{RawValue: "obsolete", Primitive: Primitive_String}, Value: &PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool}},
				},
				DefaultValue: &PrimitiveValueStmt{RawValue: int64(54), Primitive: Primitive_Int64},
			},
			err: nil,
		},
		{
			input: `5 field_name: list(string?) = ["hello", null] @[("obsolete":true)]`,
			expect: &FieldStmt{
				Index: &PrimitiveValueStmt{RawValue: int64(5), Primitive: Primitive_Int64},
				Name:  &IdentifierStmt{Lit: "field_name"},
				ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "list"}, TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Lit: "string"}, Nullable: true},
				}},
				Metadata: &MapValueStmt{
					{Key: &PrimitiveValueStmt{RawValue: "obsolete", Primitive: Primitive_String}, Value: &PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool}},
				},
				DefaultValue: &ListValueStmt{
					&PrimitiveValueStmt{RawValue: "hello", Primitive: Primitive_String},
					&PrimitiveValueStmt{RawValue: nil, Primitive: Primitive_Null},
				},
			},
			err: nil,
		},
		{
			input: `5 field_name: map(string, bool?) = [("hello": null), ("second": true)] @[("obsolete":true)]`,
			expect: &FieldStmt{
				Index: &PrimitiveValueStmt{RawValue: int64(5), Primitive: Primitive_Int64},
				Name:  &IdentifierStmt{Lit: "field_name"},
				ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "map"}, TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Lit: "string"}},
					{Ident: &IdentifierStmt{Lit: "bool"}, Nullable: true},
				}},
				Metadata: &MapValueStmt{
					{Key: &PrimitiveValueStmt{RawValue: "obsolete", Primitive: Primitive_String}, Value: &PrimitiveValueStmt{RawValue: true, Primitive: Primitive_Bool}},
				},
				DefaultValue: &MapValueStmt{
					{
						Key:   &PrimitiveValueStmt{Primitive: Primitive_String, RawValue: "hello"},
						Value: &PrimitiveValueStmt{Primitive: Primitive_Null, RawValue: nil},
					},
					{
						Key:   &PrimitiveValueStmt{Primitive: Primitive_String, RawValue: "second"},
						Value: &PrimitiveValueStmt{Primitive: Primitive_Bool, RawValue: true},
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
			expect: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "string"}},
			err:    nil,
		},
		{
			input:  `string?`,
			expect: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "string"}, Nullable: true},
			err:    nil,
		},
		{
			input: `list(string)`,
			expect: &ValueTypeStmt{
				Ident: &IdentifierStmt{Lit: "list"},
				TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Lit: "string"}},
				},
			},
			err: nil,
		},
		{
			input: `list(string,bool)`,
			expect: &ValueTypeStmt{
				Ident: &IdentifierStmt{Lit: "list"},
				TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Lit: "string"}},
					{Ident: &IdentifierStmt{Lit: "bool"}},
				},
			},
			err: nil,
		},
		{
			input: `list(string?)`,
			expect: &ValueTypeStmt{
				Ident: &IdentifierStmt{Lit: "list"},
				TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Lit: "string"}, Nullable: true},
				},
			},
			err: nil,
		},
		{
			input: `list(string?, bool?)`,
			expect: &ValueTypeStmt{
				Ident: &IdentifierStmt{Lit: "list"},
				TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Lit: "string"}, Nullable: true},
					{Ident: &IdentifierStmt{Lit: "bool"}, Nullable: true},
				},
			},
			err: nil,
		},
		{
			input: `map(int64, bool?)`,
			expect: &ValueTypeStmt{
				Ident: &IdentifierStmt{Lit: "map"},
				TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Lit: "int64"}, Nullable: false},
					{Ident: &IdentifierStmt{Lit: "bool"}, Nullable: true},
				},
			},
			err: nil,
		},
		{
			input: `list(int64)?`,
			expect: &ValueTypeStmt{
				Ident: &IdentifierStmt{Lit: "list"},
				TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Lit: "int64"}, Nullable: false},
				},
				Nullable: true,
			},
			err: nil,
		},
		{
			input: `list(int64?)?`,
			expect: &ValueTypeStmt{
				Ident: &IdentifierStmt{Lit: "list"},
				TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Lit: "int64"}, Nullable: true},
				},
				Nullable: true,
			},
			err: nil,
		},
		{
			input: `list(MyEnum)?`,
			expect: &ValueTypeStmt{
				Ident: &IdentifierStmt{Lit: "list"},
				TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Lit: "MyEnum"}},
				},
				Nullable: true,
			},
			err: nil,
		},
		{
			input: `list(another.MyEnum)?`,
			expect: &ValueTypeStmt{
				Ident: &IdentifierStmt{Lit: "list"},
				TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Alias: "another", Lit: "MyEnum"}},
				},
				Nullable: true,
			},
			err: nil,
		},
		{
			input: `map(string, another.MyEnum?)?`,
			expect: &ValueTypeStmt{
				Ident: &IdentifierStmt{Lit: "map"},
				TypeArguments: &[]*ValueTypeStmt{
					{Ident: &IdentifierStmt{Lit: "string"}},
					{Ident: &IdentifierStmt{Alias: "another", Lit: "MyEnum"}, Nullable: true},
				},
				Nullable: true,
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
				Imports: &[]*ImportStmt{
					{Path: &IdentifierStmt{Lit: "my_file.nex"}},
					{Path: &IdentifierStmt{Lit: "another.nex"}, Alias: &IdentifierStmt{Lit: "another"}},
				},
				Types: new([]*TypeStmt),
			},
			err: nil,
		},
		{
			name: "parse import with single type",
			input: `
			import:
				"my_file.nex"
				"another.nex" as another
				
			type Colors enum {
				red
				green
				blue
			}`,
			expect: &Ast{
				Imports: &[]*ImportStmt{
					{Path: &IdentifierStmt{Lit: "my_file.nex"}},
					{Path: &IdentifierStmt{Lit: "another.nex"}, Alias: &IdentifierStmt{Lit: "another"}},
				},
				Types: &[]*TypeStmt{
					{
						Name:     &IdentifierStmt{Lit: "Colors"},
						Modifier: Token_Enum,
						Fields: &[]*FieldStmt{
							{Name: &IdentifierStmt{Lit: "red"}},
							{Name: &IdentifierStmt{Lit: "green"}},
							{Name: &IdentifierStmt{Lit: "blue"}},
						},
						Documentation: new([]*CommentStmt),
					},
				},
			},
			err: nil,
		},
		{
			name: "parse import with types",
			input: `
			import:
				"my_file.nex"
				"another.nex" as another
			
			type Rectangle struct {
				x: float32
				y: float32
				color: Colors = Colors.red
			}

			@[("a":23)]
			type Colors enum {
				red
				green
				blue
			}`,
			expect: &Ast{
				Imports: &[]*ImportStmt{
					{Path: &IdentifierStmt{Lit: "my_file.nex"}},
					{Path: &IdentifierStmt{Lit: "another.nex"}, Alias: &IdentifierStmt{Lit: "another"}},
				},
				Types: &[]*TypeStmt{
					{
						Name:     &IdentifierStmt{Lit: "Rectangle"},
						Modifier: Token_Struct,
						Fields: &[]*FieldStmt{
							{
								Name:      &IdentifierStmt{Lit: "x"},
								ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "float32"}},
							},
							{
								Name:      &IdentifierStmt{Lit: "y"},
								ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "float32"}},
							},
							{
								Name:      &IdentifierStmt{Lit: "color"},
								ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "Colors"}},
								DefaultValue: &TypeValueStmt{
									TypeName: &IdentifierStmt{Lit: "Colors"},
									RawValue: &IdentifierStmt{Lit: "red"},
								},
							},
						},
						Documentation: new([]*CommentStmt),
					},
					{
						Name:     &IdentifierStmt{Lit: "Colors"},
						Modifier: Token_Enum,
						Metadata: &MapValueStmt{
							{
								Key:   &PrimitiveValueStmt{RawValue: "a", Primitive: Primitive_String},
								Value: &PrimitiveValueStmt{RawValue: int64(23), Primitive: Primitive_Int64},
							},
						},
						Fields: &[]*FieldStmt{
							{Name: &IdentifierStmt{Lit: "red"}},
							{Name: &IdentifierStmt{Lit: "green"}},
							{Name: &IdentifierStmt{Lit: "blue"}},
						},
						Documentation: new([]*CommentStmt),
					},
				},
			},
			err: nil,
		},
		{
			name: "parse with documentation",
			input: `
			import:
				"my_file.nex"
				"another.nex" as another
			
			// Rectangle is my struct
			type Rectangle struct {
				// the width
				width: float32

				// the height
				height: float32

				// the color
				color: Colors = Colors.red
			}

			// Colors is a collection of common colors
			type Colors enum {
				// Red color
				red

				// Green color
				green

				// Blue color
				blue
			}`,
			expect: &Ast{
				Imports: &[]*ImportStmt{
					{Path: &IdentifierStmt{Lit: "my_file.nex"}},
					{Path: &IdentifierStmt{Lit: "another.nex"}, Alias: &IdentifierStmt{Lit: "another"}},
				},
				Types: &[]*TypeStmt{
					{
						Name:     &IdentifierStmt{Lit: "Rectangle"},
						Modifier: Token_Struct,
						Fields: &[]*FieldStmt{
							{
								Name:      &IdentifierStmt{Lit: "width"},
								ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "float32"}},
								Documentation: &[]*CommentStmt{
									{
										Text:      "the width",
										posStart:  4,
										posEnd:    13,
										lineStart: 8,
										lineEnd:   8,
									},
								},
							},
							{
								Name:      &IdentifierStmt{Lit: "height"},
								ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "float32"}},
								Documentation: &[]*CommentStmt{
									{
										Text:      "the height",
										posStart:  4,
										posEnd:    14,
										lineStart: 11,
										lineEnd:   11,
									},
								},
							},
							{
								Name:      &IdentifierStmt{Lit: "color"},
								ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "Colors"}},
								DefaultValue: &TypeValueStmt{
									TypeName: &IdentifierStmt{Lit: "Colors"},
									RawValue: &IdentifierStmt{Lit: "red"},
								},
								Documentation: &[]*CommentStmt{
									{
										Text:      "the color",
										posStart:  4,
										posEnd:    13,
										lineStart: 14,
										lineEnd:   14,
									},
								},
							},
						},
						Documentation: &[]*CommentStmt{
							{
								Text:      "Rectangle is my struct",
								lineStart: 6,
								lineEnd:   6,
								posStart:  3,
								posEnd:    25,
							},
						},
					},
					{
						Name:     &IdentifierStmt{Lit: "Colors"},
						Modifier: Token_Enum,
						Fields: &[]*FieldStmt{
							{
								Name: &IdentifierStmt{Lit: "red"},
								Documentation: &[]*CommentStmt{
									{
										Text:      "Red color",
										posStart:  4,
										posEnd:    13,
										lineStart: 20,
										lineEnd:   20,
									},
								},
							},
							{
								Name: &IdentifierStmt{Lit: "green"},
								Documentation: &[]*CommentStmt{
									{
										Text:      "Green color",
										posStart:  4,
										posEnd:    15,
										lineStart: 23,
										lineEnd:   23,
									},
								},
							},
							{
								Name: &IdentifierStmt{Lit: "blue"},
								Documentation: &[]*CommentStmt{
									{
										Text:      "Blue color",
										posStart:  4,
										posEnd:    14,
										lineStart: 26,
										lineEnd:   26,
									},
								},
							},
						},
						Documentation: &[]*CommentStmt{
							{
								Text:      "Colors is a collection of common colors",
								lineStart: 18,
								lineEnd:   18,
								posStart:  3,
								posEnd:    42,
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			name: "parse comments that are not documentation",
			input: `
			import:
				"my_file.nex"
				"another.nex" as another
			
			/*
			multiline comments are not documentation
			*/
			type Rectangle struct {
				width: float32 // this should not be taken as documentation
				height: float32
				color: Colors = Colors.red

				// this is not documentation too
			}

			// This is a multi
			// line comment
			type Colors enum {
				red
				green
				blue
			}`,
			expect: &Ast{
				Imports: &[]*ImportStmt{
					{Path: &IdentifierStmt{Lit: "my_file.nex"}},
					{Path: &IdentifierStmt{Lit: "another.nex"}, Alias: &IdentifierStmt{Lit: "another"}},
				},
				Types: &[]*TypeStmt{
					{
						Name:     &IdentifierStmt{Lit: "Rectangle"},
						Modifier: Token_Struct,
						Fields: &[]*FieldStmt{
							{
								Name:      &IdentifierStmt{Lit: "width"},
								ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "float32"}},
							},
							{
								Name:      &IdentifierStmt{Lit: "height"},
								ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "float32"}},
							},
							{
								Name:      &IdentifierStmt{Lit: "color"},
								ValueType: &ValueTypeStmt{Ident: &IdentifierStmt{Lit: "Colors"}},
								DefaultValue: &TypeValueStmt{
									TypeName: &IdentifierStmt{Lit: "Colors"},
									RawValue: &IdentifierStmt{Lit: "red"},
								},
							},
						},
						Documentation: new([]*CommentStmt),
					},
					{
						Name:     &IdentifierStmt{Lit: "Colors"},
						Modifier: Token_Enum,
						Fields: &[]*FieldStmt{
							{
								Name: &IdentifierStmt{Lit: "red"},
							},
							{
								Name: &IdentifierStmt{Lit: "green"},
							},
							{
								Name: &IdentifierStmt{Lit: "blue"},
							},
						},
						Documentation: &[]*CommentStmt{
							{
								Text:      "This is a multi",
								posStart:  3,
								posEnd:    18,
								lineStart: 17,
								lineEnd:   17,
							},
							{
								Text:      "line comment",
								posStart:  3,
								posEnd:    15,
								lineStart: 18,
								lineEnd:   18,
							},
						},
					},
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
