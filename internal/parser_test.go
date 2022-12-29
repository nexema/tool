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
			err:   errors.New(`line: 1:1 -> strings must end with quotes (")`),
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
	} {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.parseIdentifier()
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expect, ast)
		})
	}
}
