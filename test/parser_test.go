package test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/schema_interpreter/internal"
)

func TestParse(t *testing.T) {
	type errorTestCases struct {
		input string
		err   error
	}

	for _, tt := range []errorTestCases{
		{
			input: "type MyName struct",
			err:   nil,
		},
		{
			input: "MyName",
			err:   nil,
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			parser := internal.NewParser(bytes.NewBufferString(tt.input))
			ast, err := parser.Parse()
			_ = ast
			require.Equal(t, tt.err, err)
		})
	}
}
