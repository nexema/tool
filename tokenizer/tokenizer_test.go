package tokenizer

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/token"
)

func TestTokenizerErr_IsErr(t *testing.T) {
	tests := []struct {
		name string
		err  *TokenizerErr
		is   error
		want bool
	}{
		{"ErrInvalidMultilineComment matches", NewTokenizerErr(ErrInvalidMultilineComment), ErrInvalidMultilineComment, true},
		{"ErrInvalidString matches", NewTokenizerErr(ErrInvalidString), ErrInvalidString, true},
		{"ErrUnknownToken matches", NewTokenizerErr(ErrUnknownToken), ErrUnknownToken, true},
		{"other error does not match", NewTokenizerErr(ErrUnknownToken), io.EOF, false},
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
		{"/*another comment*/", token.NewToken(token.CommentMultiline, `another comment`), NewPos(0, 19), nil},
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

func TestTokenizer_readNumber(t *testing.T) {
	tests := []struct {
		input   string
		wantPos *Pos
	}{
		{"123", NewPos(0, 3)},
		{"12.4", NewPos(0, 4)},
		{"122.4241", NewPos(0, 8)},
		{".244", NewPos(0, 4)},
		{".24.4", NewPos(0, 3)},
		{"122.2.4", NewPos(0, 5)},
		{"122..4", NewPos(0, 4)},
		{"-244", NewPos(0, 4)},
		{"-24.24", NewPos(0, 6)},
		{"-24.-24", NewPos(0, 4)},
		{"-24-", NewPos(0, 3)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(bufio.NewReader(bytes.NewBufferString(tt.input)))
			gotTok, gotPos, gotErr := tokenizer.readNumber()
			require.Nil(t, gotErr)
			require.NotNil(t, gotTok)
			require.Equal(t, tt.wantPos, gotPos)
		})
	}
}

func TestTokenizer_readString(t *testing.T) {
	tests := []struct {
		input   string
		want    *token.Token
		wantPos *Pos
		wantErr *TokenizerErr
	}{
		{`"input string"`, token.NewToken(token.String, "input string"), NewPos(0, 14), nil},
		{`"handle escape \" char"`, token.NewToken(token.String, "handle escape \" char"), NewPos(0, 23), nil},
		{`"str`, nil, nil, NewTokenizerErr(ErrInvalidString)},
		{`"s"tr"`, token.NewToken(token.String, "s"), NewPos(0, 3), nil},
		{`"other chars 漢語水平考試 😃 and numbers 1234"`, token.NewToken(token.String, "other chars 漢語水平考試 😃 and numbers 1234"), NewPos(0, 39), nil},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(bufio.NewReader(bytes.NewBufferString(tt.input)))
			gotTok, gotPos, gotErr := tokenizer.readString()
			require.Equal(t, tt.wantErr, gotErr)
			require.Equal(t, tt.want, gotTok)
			require.Equal(t, tt.wantPos, gotPos)
		})
	}
}

func TestTokenizer_readComment(t *testing.T) {
	tests := []struct {
		input   string
		want    *token.Token
		wantPos *Pos
		wantErr *TokenizerErr
	}{
		{"//simple inline comment", token.NewToken(token.Comment, "simple inline comment"), NewPos(0, 23), nil},
		{"// 漢語水平考試 😃 1234", token.NewToken(token.Comment, " 漢語水平考試 😃 1234"), NewPos(0, 16), nil},
		{"//with more //", token.NewToken(token.Comment, "with more //"), NewPos(0, 14), nil},
		{`//contains: "string inside"`, token.NewToken(token.Comment, `contains: "string inside"`), NewPos(0, 27), nil},
		{"/* multiline but inline */", token.NewToken(token.CommentMultiline, " multiline but inline "), NewPos(0, 26), nil},
		{"/* line 1 \n line \"2\"*/", token.NewToken(token.CommentMultiline, " line 1 \n line \"2\""), NewPos(0, 11, 0, 1), nil},
		{"/* error! ", nil, nil, NewTokenizerErr(ErrInvalidMultilineComment)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(bufio.NewReader(bytes.NewBufferString(tt.input)))
			gotTok, gotPos, gotErr := tokenizer.readComment()
			require.Equal(t, tt.wantErr, gotErr)
			require.Equal(t, tt.want, gotTok)
			require.Equal(t, tt.wantPos, gotPos)
		})
	}
}

func TestTokenizer_readIdentifier(t *testing.T) {
	tests := []struct {
		input   string
		want    *token.Token
		wantPos *Pos
	}{
		{"valid", token.NewToken(token.Ident, "valid"), NewPos(0, 5)},
		{"still_valid", token.NewToken(token.Ident, "still_valid"), NewPos(0, 11)},
		{"valid_too123", token.NewToken(token.Ident, "valid_too123"), NewPos(0, 12)},
		{"valid__1", token.NewToken(token.Ident, "valid__1"), NewPos(0, 8)},
		{"漢13", token.NewToken(token.Ident, "漢13"), NewPos(0, 3)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(bufio.NewReader(bytes.NewBufferString(tt.input)))
			gotTok, gotPos, gotErr := tokenizer.readIdentifier()
			require.Nil(t, gotErr)
			require.Equal(t, tt.want, gotTok)
			require.Equal(t, tt.wantPos, gotPos)
		})
	}
}
