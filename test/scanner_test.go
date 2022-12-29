package test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/schema_interpreter/internal"
)

func TestPeek(t *testing.T) {
	type errorTestCases struct {
		input   string
		token   internal.Token
		literal string
	}

	for _, tt := range []errorTestCases{
		{
			input:   "type",
			token:   internal.Token_Type,
			literal: "type",
		},
		{
			input:   "this",
			token:   internal.Token_Ident,
			literal: "this",
		},
		{
			input:   "?",
			token:   internal.Token_QuestionMark,
			literal: "?",
		},
		{
			input:   "95121",
			token:   internal.Token_Ident,
			literal: "95121",
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			scanner := internal.NewScanner(bytes.NewBufferString(tt.input))
			_, token, literal := scanner.Scan()
			require.Equal(t, tt.token, token)
			require.Equal(t, tt.literal, literal)
		})
	}
}
