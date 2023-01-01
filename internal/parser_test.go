package internal

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	type errorTestCases struct {
		input string
		err   error
		skip  bool
	}

	for _, tt := range []errorTestCases{
		{
			input: `"a random string"`,
			err:   nil,
		},
		// {
		// 	input: `
		// 	@metadata
		// 	type MyName struct {}`,
		// 	err: nil,
		// },
		// {
		// 	input: "MyName",
		// 	err:   nil,
		// },
	} {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.parseValue()
			_ = ast
			require.Equal(t, tt.err, err)
		})
	}
}

func TestParseValue(t *testing.T) {
	type testCase struct {
		input  string
		err    error
		expect baseIdentifierStmt
	}

	for _, tt := range []testCase{
		{
			input: `"a random string"`,
			expect: &identifierStmt{
				value:     "a random string",
				valueType: &valueTypeStmt{primitive: Primitive_String},
			},
			err: nil,
		},
		{
			input: `"invalid string`,
			err:   errors.New(`line 1:1 -> strings must end with quotes (")`),
		},
		{
			input: `"a random string with numbers 12345 and random characters !?¢∞¬#<"`,
			expect: &identifierStmt{
				value:     "a random string with numbers 12345 and random characters !?¢∞¬#<",
				valueType: &valueTypeStmt{primitive: Primitive_String},
			},
			err: nil,
		},
		{
			input: `"it contains a keyword: type or import inside it!"`,
			expect: &identifierStmt{
				value:     "it contains a keyword: type or import inside it!",
				valueType: &valueTypeStmt{primitive: Primitive_String},
			},
			err: nil,
		},
		{
			input: `"escaping chars? you got it \"string inside string\""`,
			expect: &identifierStmt{
				value:     `escaping chars? you got it "string inside string"`,
				valueType: &valueTypeStmt{primitive: Primitive_String},
			},
			err: nil,
		},
		{
			input: `"single escaping char \""`,
			expect: &identifierStmt{
				value:     `single escaping char "`,
				valueType: &valueTypeStmt{primitive: Primitive_String},
			},
			err: nil,
		},
		{
			input: `"multiple scaping chars \"\" and another \""`,
			expect: &identifierStmt{
				value:     `multiple scaping chars "" and another "`,
				valueType: &valueTypeStmt{primitive: Primitive_String},
			},
			err: nil,
		},
		{
			input: `123456789`,
			expect: &identifierStmt{
				value:     int64(123456789),
				valueType: &valueTypeStmt{primitive: Primitive_Int64},
			},
			err: nil,
		},
		{
			input: `12345.6789`,
			expect: &identifierStmt{
				value:     float64(12345.6789),
				valueType: &valueTypeStmt{primitive: Primitive_Float64},
			},
			err: nil,
		},
		{
			input: `true`,
			expect: &identifierStmt{
				value:     true,
				valueType: &valueTypeStmt{primitive: Primitive_Bool},
			},
			err: nil,
		},
		{
			input: `false`,
			expect: &identifierStmt{
				value:     false,
				valueType: &valueTypeStmt{primitive: Primitive_Bool},
			},
			err: nil,
		},
		{
			input: `FALSE`,
			expect: &identifierStmt{
				value:     false,
				valueType: &valueTypeStmt{primitive: Primitive_Bool},
			},
			err: nil,
		},
		{
			input: `TRUE`,
			expect: &identifierStmt{
				value:     true,
				valueType: &valueTypeStmt{primitive: Primitive_Bool},
			},
			err: nil,
		},
		{
			input:  `ta`,
			expect: nil,
			err:    errors.New("line 1:1 -> unknown primitive ta"),
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.parseValue()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ast)
		})
	}
}

func TestParseMap(t *testing.T) {
	type testCase struct {
		description string
		input       string
		err         error
		expect      *mapStmt
	}

	for _, tt := range []testCase{
		{
			description: "empty map",
			input:       `[]`,
			expect:      new(mapStmt),
			err:         nil,
		},
		{
			input: `[("key":12.43)]`,
			expect: &mapStmt{
				&mapEntryStmt{
					key: &identifierStmt{
						value:     "key",
						valueType: &valueTypeStmt{primitive: Primitive_String},
					},
					value: &identifierStmt{
						value:     float64(12.43),
						valueType: &valueTypeStmt{primitive: Primitive_Float64},
					},
				},
			},
			err: nil,
		},
		{
			description: "missing colon",
			input:       `[("key"12.43)]`,
			err:         errors.New("line 1:8 -> key-value pair must be in the format key:value"),
		},
		{
			description: "duplicated colon",
			input:       `[("key":12.43:"triple")]`,
			err:         errors.New("line 1:14 -> invalid map declaration"),
		},
		{
			description: "wrong character",
			input:       `[("key":12.43,"triple")]`,
			err:         errors.New("line 1:14 -> invalid map declaration"),
		},
		{
			description: "invalid declaration",
			input:       `[("key":12.43,type)]`,
			err:         errors.New("line 1:14 -> invalid map declaration"),
		},
		{
			description: "missing parens",
			input:       `["key":12.43]`,
			err:         errors.New(`line 1:2 -> found "\"key\"", expected "("`),
		},
		{
			description: "missing close parens",
			input:       `[("key":12.43]`,
			err:         errors.New(`line 1:14 -> invalid map declaration`),
		},
		{
			description: "missing comma separator",
			input:       `[("key":12.43)("another":true)]`,
			err:         errors.New(`line 1:15 -> map entries must be comma-separated`),
		},
		{
			input: `[("key":12.43),("second":true),("third":"yes")]`,
			expect: &mapStmt{
				&mapEntryStmt{
					key: &identifierStmt{
						value:     "key",
						valueType: &valueTypeStmt{primitive: Primitive_String},
					},
					value: &identifierStmt{
						value:     float64(12.43),
						valueType: &valueTypeStmt{primitive: Primitive_Float64},
					},
				},
				&mapEntryStmt{
					key: &identifierStmt{
						value:     "second",
						valueType: &valueTypeStmt{primitive: Primitive_String},
					},
					value: &identifierStmt{
						value:     true,
						valueType: &valueTypeStmt{primitive: Primitive_Bool},
					},
				},
				&mapEntryStmt{
					key: &identifierStmt{
						value:     "third",
						valueType: &valueTypeStmt{primitive: Primitive_String},
					},
					value: &identifierStmt{
						value:     "yes",
						valueType: &valueTypeStmt{primitive: Primitive_String},
					},
				},
			},
			err: nil,
		},
		{
			input:  `[("key":12.43),("second":true),("third":"yes"]`,
			expect: nil,
			err:    errors.New("line 1:46 -> invalid map declaration"),
		},
	} {

		testName := tt.description
		if len(tt.description) == 0 {
			testName = tt.input
		}

		t.Run(testName, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.parseMap()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ast)
		})
	}
}

func TestParseList(t *testing.T) {
	type testCase struct {
		description string
		input       string
		err         error
		expect      *listStmt
	}

	for _, tt := range []testCase{
		{
			description: "empty list",
			input:       `[]`,
			expect:      new(listStmt),
			err:         nil,
		},
		{
			description: "random elements",
			input:       `["a string", 12.32, 88, true]`,
			expect: &listStmt{
				{
					value:     "a string",
					valueType: &valueTypeStmt{primitive: Primitive_String},
				},
				{
					value:     float64(12.32),
					valueType: &valueTypeStmt{primitive: Primitive_Float64},
				},
				{
					value:     int64(88),
					valueType: &valueTypeStmt{primitive: Primitive_Int64},
				},
				{
					value:     true,
					valueType: &valueTypeStmt{primitive: Primitive_Bool},
				},
			},
			err: nil,
		},
		{
			description: "missing comma",
			input:       `["a string", 12.32, 88 true]`,
			err:         errors.New("line 1:24 -> list elements must be comma-separated"),
		},
		{
			description: "missing open bracket",
			input:       `"a string", 12.32, 88]`,
			err:         errors.New(`line 1:1 -> found "\"a string\"", expected "["`),
		},
		{
			description: "missing close bracket",
			input:       `["a string", 12.32, 88 `,
			err:         errors.New(`line 1:23 -> lists must be closed with "]"`),
		},
	} {

		testName := tt.description
		if len(tt.description) == 0 {
			testName = tt.input
		}

		t.Run(testName, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.parseList()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ast)
		})
	}
}

func TestParseType(t *testing.T) {
	t.Skip()
	type testCase struct {
		description string
		input       string
		err         error
		expect      *typeStmt
	}

	for _, tt := range []testCase{
		{
			description: "empty list",
			input:       `type MyType struct {}`,
			expect: &typeStmt{
				name:         "MyType",
				typeModifier: TypeModifier_Struct,
			},
			err: nil,
		},
	} {

		testName := tt.description
		if len(tt.description) == 0 {
			testName = tt.input
		}

		t.Run(testName, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.parseType()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ast)
		})
	}
}

func TestParseFieldType(t *testing.T) {
	type testCase struct {
		input  string
		err    error
		expect *valueTypeStmt
	}

	for _, tt := range []testCase{
		{
			input: `string`,
			expect: &valueTypeStmt{
				primitive:     Primitive_String,
				nullable:      false,
				typeArguments: nil,
			},
			err: nil,
		},
		{
			input: `string?`,
			expect: &valueTypeStmt{
				primitive:     Primitive_String,
				nullable:      true,
				typeArguments: nil,
			},
			err: nil,
		},
		{
			input: `int`,
			expect: &valueTypeStmt{
				primitive:     Primitive_Int32,
				nullable:      false,
				typeArguments: nil,
			},
			err: nil,
		},
		{
			input: `uint`,
			expect: &valueTypeStmt{
				primitive:     Primitive_Uint32,
				nullable:      false,
				typeArguments: nil,
			},
			err: nil,
		},
		{
			input: `int8?`,
			expect: &valueTypeStmt{
				primitive:     Primitive_Int8,
				nullable:      true,
				typeArguments: nil,
			},
			err: nil,
		},
		{
			input: `binary`,
			expect: &valueTypeStmt{
				primitive:     Primitive_Binary,
				nullable:      false,
				typeArguments: nil,
			},
			err: nil,
		},
		{
			input: `list(string)`,
			expect: &valueTypeStmt{
				primitive: Primitive_List,
				nullable:  false,
				typeArguments: &[]*valueTypeStmt{
					{
						primitive:     Primitive_String,
						nullable:      false,
						typeArguments: nil,
					},
				},
			},
			err: nil,
		},
		{
			input: `list(string)?`,
			expect: &valueTypeStmt{
				primitive: Primitive_List,
				nullable:  true,
				typeArguments: &[]*valueTypeStmt{
					{
						primitive:     Primitive_String,
						nullable:      false,
						typeArguments: nil,
					},
				},
			},
			err: nil,
		},
		{
			input: `list(string?)`,
			expect: &valueTypeStmt{
				primitive: Primitive_List,
				nullable:  false,
				typeArguments: &[]*valueTypeStmt{
					{
						primitive:     Primitive_String,
						nullable:      true,
						typeArguments: nil,
					},
				},
			},
			err: nil,
		},
		{
			input: `list(string?)?`,
			expect: &valueTypeStmt{
				primitive: Primitive_List,
				nullable:  true,
				typeArguments: &[]*valueTypeStmt{
					{
						primitive:     Primitive_String,
						nullable:      true,
						typeArguments: nil,
					},
				},
			},
			err: nil,
		},
		{
			input: `map(string?, float64?)?`,
			expect: &valueTypeStmt{
				primitive: Primitive_Map,
				nullable:  true,
				typeArguments: &[]*valueTypeStmt{
					{
						primitive:     Primitive_String,
						nullable:      true,
						typeArguments: nil,
					},
					{
						primitive:     Primitive_Float64,
						nullable:      true,
						typeArguments: nil,
					},
				},
			},
			err: nil,
		},
		{
			input: `map(string, boolean)`,
			expect: &valueTypeStmt{
				primitive: Primitive_Map,
				nullable:  false,
				typeArguments: &[]*valueTypeStmt{
					{
						primitive:     Primitive_String,
						nullable:      false,
						typeArguments: nil,
					},
					{
						primitive:     Primitive_Bool,
						nullable:      false,
						typeArguments: nil,
					},
				},
			},
			err: nil,
		},
		{
			input: `boo?`,
			err:   errors.New(`line 1:1 -> found "boo", expected field type`),
		},
		{
			input: `list`,
			err:   errors.New(`line 1:4 -> lists expect one type argument, given: `),
		},
		{
			input: `list(string,boolean)`,
			err:   errors.New(`line 1:12 -> found ",", expected ")"`),
		},
		{
			input: `map(string)`,
			err:   errors.New(`line 1:11 -> found ")", expected "comma"`),
		},
		{
			input: `map(string,boolean,int)`,
			err:   errors.New(`line 1:19 -> found ",", expected ")"`),
		},
		{
			input: `map(string boolean)`,
			err:   errors.New(`line 1:12 -> found "boolean", expected "comma"`),
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.parseFieldType()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ast)
		})
	}
}

func TestParseField(t *testing.T) {
	type testCase struct {
		input  string
		err    error
		expect *fieldStmt
	}

	for _, tt := range []testCase{
		{
			input: `field_one:string`,
			expect: &fieldStmt{
				name:      "field_one",
				valueType: &valueTypeStmt{primitive: Primitive_String},
				index:     0,
			},
			err: nil,
		},
		{
			input: `3 field_one:string`,
			expect: &fieldStmt{
				name:      "field_one",
				valueType: &valueTypeStmt{primitive: Primitive_String},
				index:     3,
			},
			err: nil,
		},
		{
			input: `field_one:list(float32?)`,
			expect: &fieldStmt{
				name: "field_one",
				valueType: &valueTypeStmt{
					primitive: Primitive_List,
					typeArguments: &[]*valueTypeStmt{
						{primitive: Primitive_Float32, nullable: true},
					},
				},
			},
			err: nil,
		},
		{
			input: `1 field_one:float32 = 5432.234`,
			expect: &fieldStmt{
				name:      "field_one",
				valueType: &valueTypeStmt{primitive: Primitive_Float32},
				index:     1,
				defaultValue: &identifierStmt{
					value:     float64(5432.234),
					valueType: &valueTypeStmt{primitive: Primitive_Float64},
				},
			},
			err: nil,
		},
		{
			input: `field_one:string="hello world, 12131 and \"hola\""`,
			expect: &fieldStmt{
				name:      "field_one",
				valueType: &valueTypeStmt{primitive: Primitive_String},
				defaultValue: &identifierStmt{
					value:     `hello world, 12131 and "hola"`,
					valueType: &valueTypeStmt{primitive: Primitive_String},
				},
			},
			err: nil,
		},
		{
			input: `field_one:list(boolean) = [true, false, true, true]"`,
			expect: &fieldStmt{
				name: "field_one",
				valueType: &valueTypeStmt{
					primitive:     Primitive_List,
					typeArguments: &[]*valueTypeStmt{{primitive: Primitive_Bool}},
				},
				defaultValue: &listStmt{
					{
						value:     true,
						valueType: &valueTypeStmt{primitive: Primitive_Bool},
					},
					{
						value:     false,
						valueType: &valueTypeStmt{primitive: Primitive_Bool},
					},
					{
						value:     true,
						valueType: &valueTypeStmt{primitive: Primitive_Bool},
					},
					{
						value:     true,
						valueType: &valueTypeStmt{primitive: Primitive_Bool},
					},
				},
			},
			err: nil,
		},
		{
			input: `field_one:map(string, boolean) = [("hello":true),("another":false)]"`,
			expect: &fieldStmt{
				name: "field_one",
				valueType: &valueTypeStmt{
					primitive:     Primitive_Map,
					typeArguments: &[]*valueTypeStmt{{primitive: Primitive_String}, {primitive: Primitive_Bool}},
				},
				defaultValue: &mapStmt{
					{
						key:   &identifierStmt{value: "hello", valueType: &valueTypeStmt{primitive: Primitive_String}},
						value: &identifierStmt{value: true, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
					},
					{
						key:   &identifierStmt{value: "another", valueType: &valueTypeStmt{primitive: Primitive_String}},
						value: &identifierStmt{value: false, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
					},
				},
			},
			err: nil,
		},
		{
			input: `field_one:map(string, int8?) = [("hello":null),("another":12),("negative":-2)]"`,
			expect: &fieldStmt{
				name: "field_one",
				valueType: &valueTypeStmt{
					primitive:     Primitive_Map,
					typeArguments: &[]*valueTypeStmt{{primitive: Primitive_String}, {primitive: Primitive_Int8, nullable: true}},
				},
				defaultValue: &mapStmt{
					{
						key:   &identifierStmt{value: "hello", valueType: &valueTypeStmt{primitive: Primitive_String}},
						value: &identifierStmt{value: nil, valueType: &valueTypeStmt{primitive: Primitive_Null}},
					},
					{
						key:   &identifierStmt{value: "another", valueType: &valueTypeStmt{primitive: Primitive_String}},
						value: &identifierStmt{value: int64(12), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
					},
					{
						key:   &identifierStmt{value: "negative", valueType: &valueTypeStmt{primitive: Primitive_String}},
						value: &identifierStmt{value: int64(-2), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
					},
				},
			},
			err: nil,
		},
		{
			input: `field_one:boolean = true"`,
			expect: &fieldStmt{
				name:      "field_one",
				valueType: &valueTypeStmt{primitive: Primitive_Bool},
				defaultValue: &identifierStmt{
					value:     true,
					valueType: &valueTypeStmt{primitive: Primitive_Bool},
				},
			},
			err: nil,
		},
		{
			input: `field_one:int64? = null"`,
			expect: &fieldStmt{
				name:      "field_one",
				valueType: &valueTypeStmt{primitive: Primitive_Int64, nullable: true},
				defaultValue: &identifierStmt{
					value:     nil,
					valueType: &valueTypeStmt{primitive: Primitive_Null},
				},
			},
			err: nil,
		},
		{
			input: `field_one:list(boolean)? = null"`,
			expect: &fieldStmt{
				name: "field_one",
				valueType: &valueTypeStmt{
					primitive: Primitive_List,
					nullable:  true,
					typeArguments: &[]*valueTypeStmt{
						{nullable: false, primitive: Primitive_Bool},
					},
				},
				defaultValue: &identifierStmt{
					value:     nil,
					valueType: &valueTypeStmt{primitive: Primitive_Null},
				},
			},
			err: nil,
		},
		{
			input: `field_one:list(boolean?) = [null, null, null, true]"`,
			expect: &fieldStmt{
				name: "field_one",
				valueType: &valueTypeStmt{
					primitive:     Primitive_List,
					nullable:      false,
					typeArguments: &[]*valueTypeStmt{{nullable: true, primitive: Primitive_Bool}},
				},
				defaultValue: &listStmt{
					&identifierStmt{value: nil, valueType: &valueTypeStmt{primitive: Primitive_Null}},
					&identifierStmt{value: nil, valueType: &valueTypeStmt{primitive: Primitive_Null}},
					&identifierStmt{value: nil, valueType: &valueTypeStmt{primitive: Primitive_Null}},
					&identifierStmt{value: true, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
				},
			},
			err: nil,
		},
		{
			input: `field_one:binary @[("obsolete": true),("another_key":54)]"`,
			expect: &fieldStmt{
				name:         "field_one",
				valueType:    &valueTypeStmt{primitive: Primitive_Binary},
				defaultValue: nil,
				metadata: &mapStmt{
					{
						key:   &identifierStmt{value: "obsolete", valueType: &valueTypeStmt{primitive: Primitive_String}},
						value: &identifierStmt{value: true, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
					},
					{
						key:   &identifierStmt{value: "another_key", valueType: &valueTypeStmt{primitive: Primitive_String}},
						value: &identifierStmt{value: int64(54), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
					},
				},
			},
			err: nil,
		},
		{
			input: `5 field_one:list(boolean?) = [true, null, false, true] @[("obsolete": true),("another_key":54)]"`,
			expect: &fieldStmt{
				name: "field_one",
				valueType: &valueTypeStmt{
					primitive: Primitive_List,
					typeArguments: &[]*valueTypeStmt{
						{primitive: Primitive_Bool, nullable: true},
					},
				},
				index: 5,
				defaultValue: &listStmt{
					&identifierStmt{value: true, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
					&identifierStmt{value: nil, valueType: &valueTypeStmt{primitive: Primitive_Null}},
					&identifierStmt{value: false, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
					&identifierStmt{value: true, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
				},
				metadata: &mapStmt{
					{
						key:   &identifierStmt{value: "obsolete", valueType: &valueTypeStmt{primitive: Primitive_String}},
						value: &identifierStmt{value: true, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
					},
					{
						key:   &identifierStmt{value: "another_key", valueType: &valueTypeStmt{primitive: Primitive_String}},
						value: &identifierStmt{value: int64(54), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
					},
				},
			},
			err: nil,
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.parseField()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ast)
		})
	}
}

func TestParseImport(t *testing.T) {
	type testCase struct {
		input  string
		err    error
		expect *importStmt
	}

	for _, tt := range []testCase{
		{
			input: `import "this/file.mpack"`,
			expect: &importStmt{
				src: "this/file.mpack",
			},
			err: nil,
		},
		{
			input: `import "another.mpack"`,
			expect: &importStmt{
				src: "another.mpack",
			},
			err: nil,
		},
		{
			input: `import "another.mpack" as a`,
			expect: &importStmt{
				src:   "another.mpack",
				alias: stringPointer("a"),
			},
			err: nil,
		},
		{
			input:  `import "another.mpack" as type`,
			expect: nil,
			err:    errors.New(`line 1:27 -> found "type", expected import alias`),
		},
		{
			input:  `import "another.mpack" as uint64`,
			expect: nil,
			err:    errors.New(`line 1:27 -> found "uint64", expected import alias`),
		},
		{
			input: `import "file/nested/another.mpack" AS nested`,
			expect: &importStmt{
				src:   "file/nested/another.mpack",
				alias: stringPointer("nested"),
			},
		},
		{
			input:  `import "file/nested/another.mpack" AS`,
			expect: nil,
			err:    errors.New(`line 1:37 -> found "", expected import alias`),
		},
		{
			input:  `import`,
			expect: nil,
			err:    errors.New(`line 1:6 -> found "", expected import path`),
		},
		{
			input:  `import file.mpack`,
			expect: nil,
			err:    errors.New(`line 1:8 -> found "file", expected import path`),
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.parseImport()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ast)
		})
	}
}
