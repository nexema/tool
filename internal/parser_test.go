package internal

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAll(t *testing.T) {
	type errorTestCases struct {
		description string
		input       string
		expect      *Ast
		err         error
	}

	for _, tt := range []errorTestCases{
		{
			description: "parse input",
			input: `import "my_file.mpack" AS entry
			import "another/sub.mpack"

			@[("obsolete": true)]
			type MyStruct struct {
				0 my_field: string = "hello world"
				1 another: list(boolean?) = [true, null, false, false, true]
				2 color: alias.Colors = alias.Colors.unknown
			}

			@[("obsolete": false)]
			type Colors enum {
				0 unknown
				1 red
				2 green
				3 blue
			}
			`,
			err: nil,
			expect: &Ast{
				imports: &importsStmt{
					{
						src:   "my_file.mpack",
						alias: stringPointer("entry"),
					},
					{
						src: "another/sub.mpack",
					},
				},
				types: &typesStmt{
					{
						metadata: &mapStmt{
							{
								key:   &identifierStmt{value: "obsolete", valueType: &valueTypeStmt{primitive: Primitive_String}},
								value: &identifierStmt{value: true, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
							},
						},
						name:         "MyStruct",
						typeModifier: TypeModifier_Struct,
						fields: &fieldsStmt{
							{
								index:     0,
								name:      "my_field",
								valueType: &valueTypeStmt{primitive: Primitive_String},
								defaultValue: &identifierStmt{
									value:     "hello world",
									valueType: &valueTypeStmt{primitive: Primitive_String},
								},
							},
							{
								index: 1,
								name:  "another",
								valueType: &valueTypeStmt{
									primitive:     Primitive_List,
									typeArguments: &[]*valueTypeStmt{{primitive: Primitive_Bool, nullable: true}},
								},
								defaultValue: &listStmt{
									{value: true, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
									{value: nil, valueType: &valueTypeStmt{primitive: Primitive_Null}},
									{value: false, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
									{value: false, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
									{value: true, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
								},
							},
							{
								index: 2,
								name:  "color",
								valueType: &valueTypeStmt{
									primitive:      Primitive_Type,
									customTypeName: stringPointer("alias.Colors"),
								},
								defaultValue: &customTypeIdentifierStmt{
									customTypeName: "alias.Colors",
									value:          "unknown",
								},
							},
						},
					},
					{
						metadata: &mapStmt{
							{
								key:   &identifierStmt{value: "obsolete", valueType: &valueTypeStmt{primitive: Primitive_String}},
								value: &identifierStmt{value: false, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
							},
						},
						name:         "Colors",
						typeModifier: TypeModifier_Enum,
						fields: &fieldsStmt{
							{
								index: 0,
								name:  "unknown",
							},
							{
								index: 1,
								name:  "red",
							},
							{
								index: 2,
								name:  "green",
							},
							{
								index: 3,
								name:  "blue",
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.Parse()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ast)
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
	type testCase struct {
		description string
		input       string
		err         error
		expect      *typeStmt
	}

	for _, tt := range []testCase{
		{
			description: "empty struct",
			input:       `type MyType struct {}`,
			expect: &typeStmt{
				name:         "MyType",
				typeModifier: TypeModifier_Struct,
				fields:       new(fieldsStmt),
			},
			err: nil,
		},
		{
			description: "empty union",
			input:       `type MyType union {}`,
			expect: &typeStmt{
				name:         "MyType",
				typeModifier: TypeModifier_Union,
				fields:       new(fieldsStmt),
			},
			err: nil,
		},
		{
			description: "empty enum",
			input:       `type MyType enum {}`,
			expect: &typeStmt{
				name:         "MyType",
				typeModifier: TypeModifier_Enum,
				fields:       new(fieldsStmt),
			},
			err: nil,
		},
		{
			description: "empty",
			input:       `type MyType {}`,
			expect: &typeStmt{
				name:         "MyType",
				typeModifier: TypeModifier_Struct,
				fields:       new(fieldsStmt),
			},
			err: nil,
		},
		{
			description: "all primitive types",
			input: `type MyType {
				0 bool_field: boolean
				1 string_field: string
				2 uint8_field: uint8
				3 uint16_field: uint16
				4 uint32_field: uint32
				5 uint64_field: uint64
				6 int8_field: int8
				7 int16_field: int16
				8 int32_field: int32
				9 int64_field: int64
				10 float32_field: float32
				11 float64_field: float64
				12 binary_field: binary
			}`,
			expect: &typeStmt{
				name:         "MyType",
				typeModifier: TypeModifier_Struct,
				fields: &fieldsStmt{
					{
						index:        0,
						name:         "bool_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Bool},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        1,
						name:         "string_field",
						valueType:    &valueTypeStmt{primitive: Primitive_String},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        2,
						name:         "uint8_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Uint8},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        3,
						name:         "uint16_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Uint16},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        4,
						name:         "uint32_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Uint32},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        5,
						name:         "uint64_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Uint64},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        6,
						name:         "int8_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Int8},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        7,
						name:         "int16_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Int16},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        8,
						name:         "int32_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Int32},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        9,
						name:         "int64_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Int64},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        10,
						name:         "float32_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Float32},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        11,
						name:         "float64_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Float64},
						defaultValue: nil,
						metadata:     nil,
					},
					{
						index:        12,
						name:         "binary_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Binary},
						defaultValue: nil,
						metadata:     nil,
					},
				},
			},
			err: nil,
		},
		{
			description: "all primitive types with default value",
			input: `type MyType {
				0 bool_field: boolean = true
				1 string_field: string = "hello world"
				2 uint8_field: uint8 = 12
				3 uint16_field: uint16 = 15
				4 uint32_field: uint32 = 25
				5 uint64_field: uint64 = 32
				6 int8_field: int8 = -1
				7 int16_field: int16 = -233
				8 int32_field: int32 = -25554
				9 int64_field: int64 = -256789987
				10 float32_field: float32 = 123.3245
				11 float64_field: float64 = -153.355
				12 binary_field: binary
			}`,
			expect: &typeStmt{
				name:         "MyType",
				typeModifier: TypeModifier_Struct,
				fields: &fieldsStmt{
					{
						index:        0,
						name:         "bool_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Bool},
						defaultValue: &identifierStmt{value: true, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
						metadata:     nil,
					},
					{
						index:        1,
						name:         "string_field",
						valueType:    &valueTypeStmt{primitive: Primitive_String},
						defaultValue: &identifierStmt{value: "hello world", valueType: &valueTypeStmt{primitive: Primitive_String}},
						metadata:     nil,
					},
					{
						index:        2,
						name:         "uint8_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Uint8},
						defaultValue: &identifierStmt{value: int64(12), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
						metadata:     nil,
					},
					{
						index:        3,
						name:         "uint16_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Uint16},
						defaultValue: &identifierStmt{value: int64(15), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
						metadata:     nil,
					},
					{
						index:        4,
						name:         "uint32_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Uint32},
						defaultValue: &identifierStmt{value: int64(25), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
						metadata:     nil,
					},
					{
						index:        5,
						name:         "uint64_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Uint64},
						defaultValue: &identifierStmt{value: int64(32), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
						metadata:     nil,
					},
					{
						index:        6,
						name:         "int8_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Int8},
						defaultValue: &identifierStmt{value: int64(-1), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
						metadata:     nil,
					},
					{
						index:        7,
						name:         "int16_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Int16},
						defaultValue: &identifierStmt{value: int64(-233), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
						metadata:     nil,
					},
					{
						index:        8,
						name:         "int32_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Int32},
						defaultValue: &identifierStmt{value: int64(-25554), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
						metadata:     nil,
					},
					{
						index:        9,
						name:         "int64_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Int64},
						defaultValue: &identifierStmt{value: int64(-256789987), valueType: &valueTypeStmt{primitive: Primitive_Int64}},
						metadata:     nil,
					},
					{
						index:        10,
						name:         "float32_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Float32},
						defaultValue: &identifierStmt{value: float64(123.3245), valueType: &valueTypeStmt{primitive: Primitive_Float64}},
						metadata:     nil,
					},
					{
						index:        11,
						name:         "float64_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Float64},
						defaultValue: &identifierStmt{value: float64(-153.355), valueType: &valueTypeStmt{primitive: Primitive_Float64}},
						metadata:     nil,
					},
					{
						index:        12,
						name:         "binary_field",
						valueType:    &valueTypeStmt{primitive: Primitive_Binary},
						defaultValue: nil,
						metadata:     nil,
					},
				},
			},
			err: nil,
		},
		{
			description: "type with metadata",
			input: `
				@[("obsolete": true)]
				type MyType struct {}
			`,
			expect: &typeStmt{
				name:         "MyType",
				typeModifier: TypeModifier_Struct,
				fields:       new(fieldsStmt),
				metadata: &mapStmt{
					{
						key:   &identifierStmt{value: "obsolete", valueType: &valueTypeStmt{primitive: Primitive_String}},
						value: &identifierStmt{value: true, valueType: &valueTypeStmt{primitive: Primitive_Bool}},
					},
				},
			},
			err: nil,
		},
		{
			description: "enum type with fields",
			input: `
				type Colors enum {
					0 unknown
					1 red
					2 blue
					3 green
				}
			`,
			expect: &typeStmt{
				name:         "Colors",
				typeModifier: TypeModifier_Enum,
				fields: &fieldsStmt{
					{
						index: 0,
						name:  "unknown",
					},
					{
						index: 1,
						name:  "red",
					},
					{
						index: 2,
						name:  "blue",
					},
					{
						index: 3,
						name:  "green",
					},
				},
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
		{
			input: `MyEnum`,
			expect: &valueTypeStmt{
				primitive:      Primitive_Type,
				customTypeName: stringPointer("MyEnum"),
			},
		},
		{
			input: `MyUnion?`,
			expect: &valueTypeStmt{
				primitive:      Primitive_Type,
				customTypeName: stringPointer("MyUnion"),
				nullable:       true,
			},
		},
		{
			input: `anotherfile.MyEnum?`,
			expect: &valueTypeStmt{
				primitive:      Primitive_Type,
				customTypeName: stringPointer("anotherfile.MyEnum"),
				nullable:       true,
			},
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
		input       string
		forModifier TypeModifier
		err         error
		expect      *fieldStmt
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
		{
			input:       `5 first`,
			forModifier: TypeModifier_Enum,
			expect: &fieldStmt{
				index: 5,
				name:  "first",
			},
			err: nil,
		},
		{
			input:       `unknown`,
			forModifier: TypeModifier_Enum,
			expect: &fieldStmt{
				index: 0,
				name:  "unknown",
			},
			err: nil,
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.parseField(tt.forModifier)
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
			err:    errors.New(`line 1:8 -> found "file.mpack", expected import path`),
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
