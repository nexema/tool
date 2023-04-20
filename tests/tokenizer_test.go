package tests

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
	"tomasweigenast.com/nexema/tool/internal/tokenizer"
)

func TestTokenizerErr_IsErr(t *testing.T) {
	tests := []struct {
		name string
		err  *tokenizer.TokenizerErr
		is   error
		want bool
	}{
		{"ErrInvalidMultilineComment matches", tokenizer.NewTokenizerErr(tokenizer.ErrInvalidMultilineComment), tokenizer.ErrInvalidMultilineComment, true},
		{"ErrInvalidString matches", tokenizer.NewTokenizerErr(tokenizer.ErrInvalidString), tokenizer.ErrInvalidString, true},
		{"ErrUnknownToken matches", tokenizer.NewTokenizerErr(tokenizer.ErrUnknownToken), tokenizer.ErrUnknownToken, true},
		{"other error does not match", tokenizer.NewTokenizerErr(tokenizer.ErrUnknownToken), io.EOF, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.IsErr(tt.is)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestTokenizer_Next(t *testing.T) {
	tests := []struct {
		input   string
		wantTok *token.Token
		wantPos *reference.Pos
		wantErr *tokenizer.TokenizerErr
	}{
		{"<", nil, nil, tokenizer.NewTokenizerErr(tokenizer.ErrUnknownToken, "<")},
		{"?", token.NewToken(token.QuestionMark, "?"), reference.NewPos(0, 1), nil},
		// {" ", token.NewToken(token.EOF, ""), NewPos(0, 0), nil},
		{"=", token.NewToken(token.Assign, "="), reference.NewPos(0, 1), nil},
		{":", token.NewToken(token.Colon, ":"), reference.NewPos(0, 1), nil},
		{"#", token.NewToken(token.Hash, "#"), reference.NewPos(0, 1), nil},
		{"{", token.NewToken(token.Lbrace, "{"), reference.NewPos(0, 1), nil},
		{"}", token.NewToken(token.Rbrace, "}"), reference.NewPos(0, 1), nil},
		{"[", token.NewToken(token.Lbrack, "["), reference.NewPos(0, 1), nil},
		{"]", token.NewToken(token.Rbrack, "]"), reference.NewPos(0, 1), nil},
		{"(", token.NewToken(token.Lparen, "("), reference.NewPos(0, 1), nil},
		{")", token.NewToken(token.Rparen, ")"), reference.NewPos(0, 1), nil},
		{",", token.NewToken(token.Comma, ","), reference.NewPos(0, 1), nil},
		{".", token.NewToken(token.Period, "."), reference.NewPos(0, 1), nil},
		{"// a comment", token.NewToken(token.Comment, ` a comment`), reference.NewPos(0, 12), nil},
		{"/*another comment*/", token.NewToken(token.CommentMultiline, `another comment`), reference.NewPos(0, 19), nil},
		{"12345", token.NewToken(token.Integer, "12345"), reference.NewPos(0, 5), nil},
		{"12.345", token.NewToken(token.Decimal, "12.345"), reference.NewPos(0, 6), nil},
		{`"a string"`, token.NewToken(token.String, `a string`), reference.NewPos(0, 10), nil},
		{`simple_identifier123`, token.NewToken(token.Ident, `simple_identifier123`), reference.NewPos(0, 20), nil},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := tokenizer.NewTokenizer(bufio.NewReader(bytes.NewBufferString(tt.input)))
			gotTok, gotPos, gotErr := tokenizer.Next()
			require.Equal(t, tt.wantErr, gotErr)
			require.Equal(t, tt.wantTok, gotTok)
			require.Equal(t, tt.wantPos, gotPos)
		})
	}
}
