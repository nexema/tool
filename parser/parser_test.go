package parser

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/token"
	"tomasweigenast.com/nexema/tool/tokenizer"
)

var literalKindExporter = cmp.AllowUnexported(BooleanLiteral{}, StringLiteral{}, IntLiteral{}, FloatLiteral{})

func TestParser_Consume(t *testing.T) {
	parser := newParser("abc = 123")

	// read abc. Current token should be abc, next token should be =
	err := parser.consume()
	require.Nil(t, err)
	expectTokenBuf(t, &tokenBuf{token.NewToken(token.Ident, "abc"), tokenizer.NewPos(0, 3)}, parser.currentToken)
	expectTokenBuf(t, &tokenBuf{token.NewToken(token.Assign), tokenizer.NewPos(4, 5)}, parser.nextToken)

	// current token should be =, next token should be 123
	err = parser.consume()
	require.Nil(t, err)
	expectTokenBuf(t, &tokenBuf{token.NewToken(token.Assign), tokenizer.NewPos(4, 5)}, parser.currentToken)
	expectTokenBuf(t, &tokenBuf{token.NewToken(token.Integer, "123"), tokenizer.NewPos(6, 9)}, parser.nextToken)

	// current token should be 123, next token should be eof
	err = parser.consume()
	require.Nil(t, err)
	expectTokenBuf(t, &tokenBuf{token.NewToken(token.Integer, "123"), tokenizer.NewPos(6, 9)}, parser.currentToken)
	expectTokenBuf(t, nil, parser.nextToken)
}

func TestParser_Next(t *testing.T) {
	tests := []struct {
		input                     string
		want                      []tokenBuf
		wantAnnotationsOrComments []annotationOrComment
	}{
		{
			input:                     "25 = true",
			wantAnnotationsOrComments: nil,
			want: []tokenBuf{
				{token.NewToken(token.Integer, "25"), tokenizer.NewPos(0, 2)},
				{token.NewToken(token.Assign), tokenizer.NewPos(3, 4)},
				{token.NewToken(token.Ident, "true"), tokenizer.NewPos(5, 9)},
			},
		},
		{
			input: "nice // hello",
			wantAnnotationsOrComments: []annotationOrComment{
				{comment: &CommentStmt{
					Token: *token.NewToken(token.Comment, " hello"),
					Pos:   *tokenizer.NewPos(5, 13),
				}},
			},
			want: []tokenBuf{{token.NewToken(token.Ident, "nice"), tokenizer.NewPos(0, 4)}},
		},
		{
			input: `12.42 /*
		    another comment
            */ true`,
			wantAnnotationsOrComments: nil,
			want: []tokenBuf{
				{token.NewToken(token.Decimal, "12.42"), tokenizer.NewPos(0, 5)},
				{token.NewToken(token.Ident, "true"), tokenizer.NewPos(15, 19, 2, 2)},
			},
		},
		{
			input: `#obsolete = true nice`,
			wantAnnotationsOrComments: []annotationOrComment{
				{annotation: &AnnotationStmt{
					Token: *token.NewToken(token.Hash),
					Assigment: AssignStmt{
						Token: *token.NewToken(token.Assign),
						Left: IdentStmt{
							Token: *token.NewToken(token.Ident, "obsolete"),
							Pos:   *tokenizer.NewPos(1, 9),
						},
						Right: LiteralStmt{
							Token: *token.NewToken(token.Ident, "true"),
							Kind:  BooleanLiteral{true},
							Pos:   *tokenizer.NewPos(12, 16),
						},
						Pos: *tokenizer.NewPos(1, 16),
					},
					Pos: *tokenizer.NewPos(0, 16),
				}},
			},
			want: []tokenBuf{{token.NewToken(token.Ident, "nice"), tokenizer.NewPos(17, 21)}},
		},
	}

	readToEnd := func(t *testing.T, parser *Parser) []tokenBuf {
		out := make([]tokenBuf, 0)
		for {
			parser.next()
			if !parser.errors.IsEmpty() {
				t.Fatalf("error readling all tokens: %v", parser.errors.Display())
			}

			if parser.eof {
				break
			}

			out = append(out, *parser.currentToken)
		}

		return out
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := newParser(tt.input)

			tokens := readToEnd(t, parser)
			if diff := cmp.Diff(tt.want, tokens, cmp.AllowUnexported(tokenBuf{})); diff != "" {
				t.Errorf("TestParser_Next: %s() mismatch (-want +got):\n%s", tt.input, diff)
			}

			var annotationsOrComments []annotationOrComment = nil
			if parser.annotationsOrComments.Len() > 0 {
				annotationsOrComments = make([]annotationOrComment, 0)
				parser.annotationsOrComments.Ascend(func(k int, v *[]annotationOrComment) {
					annotationsOrComments = append(annotationsOrComments, *v...)
				})
			}

			if diff := cmp.Diff(tt.wantAnnotationsOrComments, annotationsOrComments, cmp.AllowUnexported(annotationOrComment{}, BooleanLiteral{}, StringLiteral{}, IntLiteral{}, FloatLiteral{})); diff != "" {
				t.Errorf("TestParser_Next: %s() mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestParser_ParseLiteral(t *testing.T) {
	tests := []struct {
		input   string
		want    *LiteralStmt
		wantErr *ParserError
	}{
		{"true", &LiteralStmt{*token.NewToken(token.Ident, "true"), BooleanLiteral{true}, *tokenizer.NewPos(0, 4)}, nil},
		{"false", &LiteralStmt{*token.NewToken(token.Ident, "false"), BooleanLiteral{false}, *tokenizer.NewPos(0, 5)}, nil},
		{"12.53", &LiteralStmt{*token.NewToken(token.Decimal, "12.53"), FloatLiteral{12.53}, *tokenizer.NewPos(0, 5)}, nil},
		{".53", &LiteralStmt{*token.NewToken(token.Decimal, ".53"), FloatLiteral{.53}, *tokenizer.NewPos(0, 3)}, nil},
		{"12", &LiteralStmt{*token.NewToken(token.Integer, "12"), IntLiteral{12}, *tokenizer.NewPos(0, 2)}, nil},
		{`"my string"`, &LiteralStmt{*token.NewToken(token.String, "my string"), StringLiteral{"my string"}, *tokenizer.NewPos(0, 11)}, nil},
		{"my_type", nil, NewParserErr(ErrInvalidLiteral{*token.NewToken(token.Ident, "my_type")}, *tokenizer.NewPos(0, 7))},
		{`["hello", true, 12.343, .12, 98, false]`, &LiteralStmt{
			Token: *token.NewToken(token.List),
			Kind: ListLiteral{
				{*token.NewToken(token.String, "hello"), StringLiteral{"hello"}, *tokenizer.NewPos(1, 8)},
				{*token.NewToken(token.Ident, "true"), BooleanLiteral{true}, *tokenizer.NewPos(10, 14)},
				{*token.NewToken(token.Decimal, "12.343"), FloatLiteral{12.343}, *tokenizer.NewPos(16, 22)},
				{*token.NewToken(token.Decimal, ".12"), FloatLiteral{.12}, *tokenizer.NewPos(24, 27)},
				{*token.NewToken(token.Integer, "98"), IntLiteral{98}, *tokenizer.NewPos(29, 31)},
				{*token.NewToken(token.Ident, "false"), BooleanLiteral{false}, *tokenizer.NewPos(33, 38)},
			},
			Pos: *tokenizer.NewPos(0, 39),
		}, nil},
		{`["hello", true,]`, &LiteralStmt{
			Token: *token.NewToken(token.List),
			Kind: ListLiteral{
				{*token.NewToken(token.String, "hello"), StringLiteral{"hello"}, *tokenizer.NewPos(1, 8)},
				{*token.NewToken(token.Ident, "true"), BooleanLiteral{true}, *tokenizer.NewPos(10, 14)},
			},
			Pos: *tokenizer.NewPos(0, 16),
		}, nil},
		{`["hello", true,`, nil, NewParserErr(ErrUnexpectedEOF{}, *tokenizer.NewPos(15, 15))},
		{`["hello", true`, nil, NewParserErr(ErrUnexpectedEOF{}, *tokenizer.NewPos(14, 14))},
		{`["hello" true`, nil, NewParserErr(ErrUnexpectedToken{token.Rbrack, *token.NewToken(token.Ident, "true")}, *tokenizer.NewPos(9, 13))},
		{`{"str": "yes","str_2": true, "str_3": 12.3, "str_4": 98, "str_5": false, 42: "str"}`, &LiteralStmt{
			Token: *token.NewToken(token.Map),
			Pos:   *tokenizer.NewPos(0, 83),
			Kind: MapLiteral{
				{
					Key:   LiteralStmt{*token.NewToken(token.String, "str"), StringLiteral{"str"}, *tokenizer.NewPos(1, 6)},
					Value: LiteralStmt{*token.NewToken(token.String, "yes"), StringLiteral{"yes"}, *tokenizer.NewPos(8, 13)},
				},
				{
					Key:   LiteralStmt{*token.NewToken(token.String, "str_2"), StringLiteral{"str_2"}, *tokenizer.NewPos(14, 21)},
					Value: LiteralStmt{*token.NewToken(token.Ident, "true"), BooleanLiteral{true}, *tokenizer.NewPos(23, 27)},
				},
				{
					Key:   LiteralStmt{*token.NewToken(token.String, "str_3"), StringLiteral{"str_3"}, *tokenizer.NewPos(29, 36)},
					Value: LiteralStmt{*token.NewToken(token.Decimal, "12.3"), FloatLiteral{12.3}, *tokenizer.NewPos(38, 42)},
				},
				{
					Key:   LiteralStmt{*token.NewToken(token.String, "str_4"), StringLiteral{"str_4"}, *tokenizer.NewPos(44, 51)},
					Value: LiteralStmt{*token.NewToken(token.Integer, "98"), IntLiteral{98}, *tokenizer.NewPos(53, 55)},
				},
				{
					Key:   LiteralStmt{*token.NewToken(token.String, "str_5"), StringLiteral{"str_5"}, *tokenizer.NewPos(57, 64)},
					Value: LiteralStmt{*token.NewToken(token.Ident, "false"), BooleanLiteral{false}, *tokenizer.NewPos(66, 71)},
				},
				{
					Key:   LiteralStmt{*token.NewToken(token.Integer, "42"), IntLiteral{42}, *tokenizer.NewPos(73, 75)},
					Value: LiteralStmt{*token.NewToken(token.String, "str"), StringLiteral{"str"}, *tokenizer.NewPos(77, 82)},
				},
			},
		}, nil},
		{`{"str": "yes","str_2": true,}`, &LiteralStmt{
			Token: *token.NewToken(token.Map),
			Pos:   *tokenizer.NewPos(0, 29),
			Kind: MapLiteral{
				{
					Key:   LiteralStmt{*token.NewToken(token.String, "str"), StringLiteral{"str"}, *tokenizer.NewPos(1, 6)},
					Value: LiteralStmt{*token.NewToken(token.String, "yes"), StringLiteral{"yes"}, *tokenizer.NewPos(8, 13)},
				},
				{
					Key:   LiteralStmt{*token.NewToken(token.String, "str_2"), StringLiteral{"str_2"}, *tokenizer.NewPos(14, 21)},
					Value: LiteralStmt{*token.NewToken(token.Ident, "true"), BooleanLiteral{true}, *tokenizer.NewPos(23, 27)},
				},
			},
		}, nil},
		{`{"str": "yes","str_2": true,`, nil, NewParserErr(ErrUnexpectedEOF{}, *tokenizer.NewPos(28, 28))},
		{`{"str": "yes","str_2": true`, nil, NewParserErr(ErrUnexpectedEOF{}, *tokenizer.NewPos(27, 27))},
		{`{"str": "yes","str_2":`, nil, NewParserErr(ErrUnexpectedEOF{}, *tokenizer.NewPos(22, 22))},
		{`{"str": "yes","str_2" true`, nil, NewParserErr(ErrUnexpectedToken{token.Colon, *token.NewToken(token.Ident, "true")}, *tokenizer.NewPos(22, 26))},
		{`{"str": "yes" "str_2": true`, nil, NewParserErr(ErrUnexpectedToken{token.Rbrace, *token.NewToken(token.String, "str_2")}, *tokenizer.NewPos(14, 21))},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := newParser(tt.input)
			parser.Reset()

			stmt := parser.parseLiteral()
			if tt.wantErr == nil {
				require.Empty(t, parser.errors)
			} else {
				require.NotEmpty(t, parser.errors)
				require.Equal(t, *tt.wantErr, *(*parser.errors)[0])
			}

			if diff := cmp.Diff(tt.want, stmt, literalKindExporter); diff != "" {
				t.Errorf("TestParser_ParseLiteral: %s -> mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestParser_ParseDecl(t *testing.T) {
	tests := []struct {
		input   string
		want    *DeclStmt
		wantErr *ParserError
	}{
		{"string", &DeclStmt{*token.NewToken(token.Ident, "string"), *tokenizer.NewPos(0, 6), nil, nil, false}, nil},
		{"string?", &DeclStmt{*token.NewToken(token.Ident, "string"), *tokenizer.NewPos(0, 6), nil, nil, true}, nil},
		{"MyEnum", &DeclStmt{*token.NewToken(token.Ident, "MyEnum"), *tokenizer.NewPos(0, 6), nil, nil, false}, nil},
		{"MyEnum?", &DeclStmt{*token.NewToken(token.Ident, "MyEnum"), *tokenizer.NewPos(0, 6), nil, nil, true}, nil},
		{"package.MyEnum", &DeclStmt{*token.NewToken(token.Ident, "MyEnum"), *tokenizer.NewPos(0, 14), nil, &IdentStmt{*token.NewToken(token.Ident, "package"), *tokenizer.NewPos(0, 7)}, false}, nil},
		{"package.MyEnum?", &DeclStmt{*token.NewToken(token.Ident, "MyEnum"), *tokenizer.NewPos(0, 14), nil, &IdentStmt{*token.NewToken(token.Ident, "package"), *tokenizer.NewPos(0, 7)}, true}, nil},
		{"list(bool)", &DeclStmt{*token.NewToken(token.Ident), *tokenizer.NewPos(0, 10), []DeclStmt{
			{*token.NewToken(token.Ident, "bool"), *tokenizer.NewPos(5, 9), nil, nil, false},
		}, nil, false}, nil},
		{"list(bool)?", &DeclStmt{*token.NewToken(token.Ident), *tokenizer.NewPos(0, 10), []DeclStmt{
			{*token.NewToken(token.Ident, "bool"), *tokenizer.NewPos(5, 9), nil, nil, false},
		}, nil, true}, nil},
		{"list(bool?)", &DeclStmt{*token.NewToken(token.Ident), *tokenizer.NewPos(0, 11), []DeclStmt{
			{*token.NewToken(token.Ident, "bool"), *tokenizer.NewPos(5, 9), nil, nil, true},
		}, nil, false}, nil},
		{"list(bool?)?", &DeclStmt{*token.NewToken(token.Ident), *tokenizer.NewPos(0, 11), []DeclStmt{
			{*token.NewToken(token.Ident, "bool"), *tokenizer.NewPos(5, 9), nil, nil, true},
		}, nil, true}, nil},
		{"list(package.MyEnum)", &DeclStmt{*token.NewToken(token.Ident), *tokenizer.NewPos(0, 20), []DeclStmt{
			{*token.NewToken(token.Ident, "MyEnum"), *tokenizer.NewPos(5, 19), nil, &IdentStmt{*token.NewToken(token.Ident, "package"), *tokenizer.NewPos(5, 12)}, false},
		}, nil, false}, nil},
		{"map(string?, package.MyEnum?)?", &DeclStmt{*token.NewToken(token.Ident), *tokenizer.NewPos(0, 29), []DeclStmt{
			{*token.NewToken(token.Ident, "string"), *tokenizer.NewPos(4, 10), nil, nil, true},
			{*token.NewToken(token.Ident, "MyEnum"), *tokenizer.NewPos(13, 27), nil, &IdentStmt{*token.NewToken(token.Ident, "package"), *tokenizer.NewPos(13, 20)}, true},
		}, nil, true}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := newParser(tt.input)
			parser.Reset()

			stmt := parser.parseDeclStmt(true)
			if tt.wantErr == nil {
				require.Empty(t, parser.errors)
			} else {
				require.NotEmpty(t, parser.errors)
				require.Equal(t, *tt.wantErr, *(*parser.errors)[0])
			}

			if diff := cmp.Diff(tt.want, stmt, literalKindExporter); diff != "" {
				t.Errorf("TestParser_ParseDecl: %s -> mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestParser_ParseAssign(t *testing.T) {
	tests := []struct {
		input   string
		want    *AssignStmt
		wantErr *ParserError
	}{
		{"my_field = true", &AssignStmt{
			Token: *token.NewToken(token.Assign),
			Pos:   *tokenizer.NewPos(0, 15),
			Left: IdentStmt{
				Token: *token.NewToken(token.Ident, "my_field"),
				Pos:   *tokenizer.NewPos(0, 8),
			},
			Right: LiteralStmt{
				Token: *token.NewToken(token.Ident, "true"),
				Kind:  BooleanLiteral{true},
				Pos:   *tokenizer.NewPos(11, 15),
			},
		}, nil},
		{`my_field = "hello"`, &AssignStmt{
			Token: *token.NewToken(token.Assign),
			Pos:   *tokenizer.NewPos(0, 18),
			Left: IdentStmt{
				Token: *token.NewToken(token.Ident, "my_field"),
				Pos:   *tokenizer.NewPos(0, 8),
			},
			Right: LiteralStmt{
				Token: *token.NewToken(token.String, "hello"),
				Kind:  StringLiteral{"hello"},
				Pos:   *tokenizer.NewPos(11, 18),
			},
		}, nil},
		{"my_field true", nil, NewParserErr(ErrUnexpectedToken{token.Assign, *token.NewToken(token.Ident, "true")}, *tokenizer.NewPos(9, 13))},
		{"my_field =", nil, NewParserErr(ErrExpectedLiteral{*token.Token_EOF}, *tokenizer.NewPos(10, 10))},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := newParser(tt.input)
			parser.Reset()

			stmt := parser.parseAssignStmt()
			if tt.wantErr == nil {
				require.Empty(t, parser.errors)
			} else {
				require.NotEmpty(t, parser.errors)
				require.Equal(t, *tt.wantErr, *(*parser.errors)[0])
			}

			if diff := cmp.Diff(tt.want, stmt, literalKindExporter); diff != "" {
				t.Errorf("TestParser_ParseAssign: %s -> mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestParser_ParseAnnotation(t *testing.T) {
	tests := []struct {
		input   string
		want    *AnnotationStmt
		wantErr *ParserError
	}{
		{"#my_field = true", &AnnotationStmt{
			Token: *token.NewToken(token.Hash),
			Pos:   *tokenizer.NewPos(0, 16),
			Assigment: AssignStmt{
				Token: *token.NewToken(token.Assign),
				Pos:   *tokenizer.NewPos(1, 16),
				Left: IdentStmt{
					Token: *token.NewToken(token.Ident, "my_field"),
					Pos:   *tokenizer.NewPos(1, 9),
				},
				Right: LiteralStmt{
					Token: *token.NewToken(token.Ident, "true"),
					Kind:  BooleanLiteral{true},
					Pos:   *tokenizer.NewPos(12, 16),
				},
			},
		}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := newParser(tt.input)
			parser.Reset()
			parser.next()

			stmt := parser.parseAnnotationStmt()
			if tt.wantErr == nil {
				require.Empty(t, parser.errors)
			} else {
				require.NotEmpty(t, parser.errors)
				require.Equal(t, *tt.wantErr, *(*parser.errors)[0])
			}

			if diff := cmp.Diff(tt.want, stmt, literalKindExporter); diff != "" {
				t.Errorf("TestParser_ParseAnnotation: %s -> mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestParser_ParseField(t *testing.T) {
	tests := []struct {
		input                       string
		inputAnnotationOrStatements map[int][]annotationOrComment
		isEnum                      bool
		want                        *FieldStmt
		wantErr                     *ParserError
	}{
		{},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := newParser(tt.input)
			parser.Reset()
			parser.next()

			stmt := parser.parseFieldStmt(tt.isEnum)
			if tt.wantErr == nil {
				require.Empty(t, parser.errors)
			} else {
				require.NotEmpty(t, parser.errors)
				require.Equal(t, *tt.wantErr, *(*parser.errors)[0])
			}

			if diff := cmp.Diff(tt.want, stmt, literalKindExporter); diff != "" {
				t.Errorf("TestParser_ParseField: %s -> mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func expectTokenBuf(t *testing.T, expected, given *tokenBuf) {
	if expected == nil {
		if given != nil {
			t.Fatalf("given tokenBuf expected to be nil, got %v instead", given)
		}
	} else {
		require.NotNil(t, given, "tokenBuf expected to be not nil")
		require.Equal(t, *expected, *given)
	}
}

func newParser(i string) *Parser {
	p := NewParser(bytes.NewBufferString(i), nil)
	p.Reset()

	return p
}
