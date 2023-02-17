package tokenizer

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/token"
)

func TestTokenizer_Next(t *testing.T) {
	tests := []struct {
		input   string
		wantTok *token.Token
		wantPos *Pos
		wantErr *TokenizerErr
	}{
		{"<", nil, nil, NewTokenizerErr(ErrUnknownToken, "<")},
		{"?", token.NewToken(token.QuestionMark, "?"), NewPos(0, 1), nil},
		// {" ", token.NewToken(token.EOF, ""), NewPos(0, 0), nil},
		{"=", token.NewToken(token.Assign, "="), NewPos(0, 1), nil},
		{":", token.NewToken(token.Colon, ":"), NewPos(0, 1), nil},
		{"#", token.NewToken(token.Hash, "#"), NewPos(0, 1), nil},
		{"{", token.NewToken(token.Lbrace, "{"), NewPos(0, 1), nil},
		{"}", token.NewToken(token.Rbrace, "}"), NewPos(0, 1), nil},
		{"[", token.NewToken(token.Lbrack, "["), NewPos(0, 1), nil},
		{"]", token.NewToken(token.Rbrack, "]"), NewPos(0, 1), nil},
		{"(", token.NewToken(token.Lparen, "("), NewPos(0, 1), nil},
		{")", token.NewToken(token.Rparen, ")"), NewPos(0, 1), nil},
		{",", token.NewToken(token.Comma, ","), NewPos(0, 1), nil},
		{".", token.NewToken(token.Period, "."), NewPos(0, 1), nil},
		{"// a comment", token.NewToken(token.Comment, ` a comment`), NewPos(0, 12), nil},
		{"/*another comment*/", token.NewToken(token.Comment, `another comment`), NewPos(0, 19), nil},
		{"12345", token.NewToken(token.Integer, "12345"), NewPos(0, 5), nil},
		{"12.345", token.NewToken(token.Decimal, "12.345"), NewPos(0, 6), nil},
		{`"a string"`, token.NewToken(token.String, `a string`), NewPos(0, 10), nil},
		{`simple_identifier123`, token.NewToken(token.Ident, `simple_identifier123`), NewPos(0, 20), nil},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(bufio.NewReader(bytes.NewBufferString(tt.input)))
			gotTok, gotPos, gotErr := tokenizer.Next()
			require.Equal(t, tt.wantErr, gotErr)
			require.Equal(t, tt.wantTok, gotTok)
			require.Equal(t, tt.wantPos, gotPos)
		})
	}
}

func TestTokenizer_next(t *testing.T) {
	type fields struct {
		reader      *bufio.Reader
		ch          rune
		currentPos  int
		currentLine int
	}
	tests := []struct {
		name   string
		fields fields
		want   rune
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := &Tokenizer{
				reader:      tt.fields.reader,
				ch:          tt.fields.ch,
				currentPos:  tt.fields.currentPos,
				currentLine: tt.fields.currentLine,
			}
			if got := self.next(); got != tt.want {
				t.Errorf("Tokenizer.next() = %v, want %v", got, tt.want)
			}
		})
	}
}
