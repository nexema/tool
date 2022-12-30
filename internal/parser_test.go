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
			ast, err := parser.parseIdentifier()
			_ = ast
			require.Equal(t, tt.err, err)
		})
	}
}

func TestParseIdentifier(t *testing.T) {
	type testCase struct {
		input  string
		err    error
		expect *identifierStmt
	}

	for _, tt := range []testCase{
		{
			input: `"a random string"`,
			expect: &identifierStmt{
				value:     "a random string",
				valueType: Primitive_String,
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
				valueType: Primitive_String,
			},
			err: nil,
		},
		{
			input: `"it contains a keyword: type or import inside it!"`,
			expect: &identifierStmt{
				value:     "it contains a keyword: type or import inside it!",
				valueType: Primitive_String,
			},
			err: nil,
		},
		{
			input: `"escaping chars? you got it \"string inside string\""`,
			expect: &identifierStmt{
				value:     `escaping chars? you got it "string inside string"`,
				valueType: Primitive_String,
			},
			err: nil,
		},
		{
			input: `"single escaping char \""`,
			expect: &identifierStmt{
				value:     `single escaping char "`,
				valueType: Primitive_String,
			},
			err: nil,
		},
		{
			input: `"multiple scaping chars \"\" and another \""`,
			expect: &identifierStmt{
				value:     `multiple scaping chars "" and another "`,
				valueType: Primitive_String,
			},
			err: nil,
		},
		{
			input: `123456789`,
			expect: &identifierStmt{
				value:     int64(123456789),
				valueType: Primitive_Int64,
			},
			err: nil,
		},
		{
			input: `12345.6789`,
			expect: &identifierStmt{
				value:     float64(12345.6789),
				valueType: Primitive_Float64,
			},
			err: nil,
		},
		{
			input: `true`,
			expect: &identifierStmt{
				value:     true,
				valueType: Primitive_Bool,
			},
			err: nil,
		},
		{
			input: `false`,
			expect: &identifierStmt{
				value:     false,
				valueType: Primitive_Bool,
			},
			err: nil,
		},
		{
			input: `FALSE`,
			expect: &identifierStmt{
				value:     false,
				valueType: Primitive_Bool,
			},
			err: nil,
		},
		{
			input: `TRUE`,
			expect: &identifierStmt{
				value:     true,
				valueType: Primitive_Bool,
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
			ast, err := parser.parseIdentifier()
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
						valueType: Primitive_String,
					},
					value: &identifierStmt{
						value:     float64(12.43),
						valueType: Primitive_Float64,
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
						valueType: Primitive_String,
					},
					value: &identifierStmt{
						value:     float64(12.43),
						valueType: Primitive_Float64,
					},
				},
				&mapEntryStmt{
					key: &identifierStmt{
						value:     "second",
						valueType: Primitive_String,
					},
					value: &identifierStmt{
						value:     true,
						valueType: Primitive_Bool,
					},
				},
				&mapEntryStmt{
					key: &identifierStmt{
						value:     "third",
						valueType: Primitive_String,
					},
					value: &identifierStmt{
						value:     "yes",
						valueType: Primitive_String,
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
					valueType: Primitive_String,
				},
				{
					value:     float64(12.32),
					valueType: Primitive_Float64,
				},
				{
					value:     int64(88),
					valueType: Primitive_Int64,
				},
				{
					value:     true,
					valueType: Primitive_Bool,
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
