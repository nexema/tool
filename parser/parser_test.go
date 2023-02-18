package parser

import (
	"bytes"
	"strings"
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
				parser.annotationsOrComments.Scan(func(key int, value *[]annotationOrComment) bool {
					annotationsOrComments = append(annotationsOrComments, *value...)
					return true
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
		{`["hello", true,`, nil, NewParserErr(ErrExpectedLiteral{*token.Token_EOF}, *tokenizer.NewPos(15, 15))},
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
		{`{"str": "yes","str_2": true,`, nil, NewParserErr(ErrExpectedLiteral{*token.Token_EOF}, *tokenizer.NewPos(28, 28))},
		{`{"str": "yes","str_2": true`, nil, NewParserErr(ErrUnexpectedEOF{}, *tokenizer.NewPos(27, 27))},
		{`{"str": "yes","str_2":`, nil, NewParserErr(ErrExpectedLiteral{*token.Token_EOF}, *tokenizer.NewPos(22, 22))},
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
		{"list(bool)", &DeclStmt{*token.NewToken(token.Ident, "list"), *tokenizer.NewPos(0, 10), []DeclStmt{
			{*token.NewToken(token.Ident, "bool"), *tokenizer.NewPos(5, 9), nil, nil, false},
		}, nil, false}, nil},
		{"list(bool)?", &DeclStmt{*token.NewToken(token.Ident, "list"), *tokenizer.NewPos(0, 10), []DeclStmt{
			{*token.NewToken(token.Ident, "bool"), *tokenizer.NewPos(5, 9), nil, nil, false},
		}, nil, true}, nil},
		{"list(bool?)", &DeclStmt{*token.NewToken(token.Ident, "list"), *tokenizer.NewPos(0, 11), []DeclStmt{
			{*token.NewToken(token.Ident, "bool"), *tokenizer.NewPos(5, 9), nil, nil, true},
		}, nil, false}, nil},
		{"list(bool?)?", &DeclStmt{*token.NewToken(token.Ident, "list"), *tokenizer.NewPos(0, 11), []DeclStmt{
			{*token.NewToken(token.Ident, "bool"), *tokenizer.NewPos(5, 9), nil, nil, true},
		}, nil, true}, nil},
		{"list(package.MyEnum)", &DeclStmt{*token.NewToken(token.Ident, "list"), *tokenizer.NewPos(0, 20), []DeclStmt{
			{*token.NewToken(token.Ident, "MyEnum"), *tokenizer.NewPos(5, 19), nil, &IdentStmt{*token.NewToken(token.Ident, "package"), *tokenizer.NewPos(5, 12)}, false},
		}, nil, false}, nil},
		{"map(string?, package.MyEnum?)?", &DeclStmt{*token.NewToken(token.Ident, "map"), *tokenizer.NewPos(0, 29), []DeclStmt{
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
		setLine                     int
		inputAnnotationOrStatements map[int]*[]annotationOrComment
		isEnum                      bool
		want                        *FieldStmt
		wantErr                     *ParserError
	}{
		{
			input:                       "0 my_field string",
			inputAnnotationOrStatements: nil,
			isEnum:                      false,
			want: &FieldStmt{
				Index:         &IdentStmt{*token.NewToken(token.Integer, "0"), *tokenizer.NewPos(0, 1)},
				Name:          IdentStmt{*token.NewToken(token.Ident, "my_field"), *tokenizer.NewPos(2, 10)},
				ValueType:     &DeclStmt{*token.NewToken(token.Ident, "string"), *tokenizer.NewPos(11, 17), nil, nil, false},
				Documentation: nil,
				Annotations:   nil,
			},
			wantErr: nil,
		},
		{
			input:                       "my_field string",
			inputAnnotationOrStatements: nil,
			isEnum:                      false,
			want: &FieldStmt{
				Index:         nil,
				Name:          IdentStmt{*token.NewToken(token.Ident, "my_field"), *tokenizer.NewPos(0, 8)},
				ValueType:     &DeclStmt{*token.NewToken(token.Ident, "string"), *tokenizer.NewPos(9, 15), nil, nil, false},
				Documentation: nil,
				Annotations:   nil,
			},
			wantErr: nil,
		},
		{
			input:                       "1 my_field list(string)?",
			inputAnnotationOrStatements: nil,
			isEnum:                      false,
			want: &FieldStmt{
				Index: &IdentStmt{*token.NewToken(token.Integer, "1"), *tokenizer.NewPos(0, 1)},
				Name:  IdentStmt{*token.NewToken(token.Ident, "my_field"), *tokenizer.NewPos(2, 10)},
				ValueType: &DeclStmt{*token.NewToken(token.Ident, "list"), *tokenizer.NewPos(11, 23), []DeclStmt{
					{*token.NewToken(token.Ident, "string"), *tokenizer.NewPos(16, 22), nil, nil, false},
				}, nil, true},
				Documentation: nil,
				Annotations:   nil,
			},
			wantErr: nil,
		},
		{
			input:                       "1 my_field map(string,bool)",
			inputAnnotationOrStatements: nil,
			isEnum:                      false,
			want: &FieldStmt{
				Index: &IdentStmt{*token.NewToken(token.Integer, "1"), *tokenizer.NewPos(0, 1)},
				Name:  IdentStmt{*token.NewToken(token.Ident, "my_field"), *tokenizer.NewPos(2, 10)},
				ValueType: &DeclStmt{*token.NewToken(token.Ident, "map"), *tokenizer.NewPos(11, 27), []DeclStmt{
					{*token.NewToken(token.Ident, "string"), *tokenizer.NewPos(15, 21), nil, nil, false},
					{*token.NewToken(token.Ident, "bool"), *tokenizer.NewPos(22, 26), nil, nil, false},
				}, nil, false},
				Documentation: nil,
				Annotations:   nil,
			},
			wantErr: nil,
		},
		{
			input:                       "red",
			inputAnnotationOrStatements: nil,
			isEnum:                      true,
			want: &FieldStmt{
				Name:          IdentStmt{*token.NewToken(token.Ident, "red"), *tokenizer.NewPos(0, 3)},
				Documentation: nil,
				Annotations:   nil,
			},
			wantErr: nil,
		},
		{
			input:   "my_field 22 map(string,bool)",
			wantErr: NewParserErr(ErrExpectedIdentifier{*token.NewToken(token.Integer, "22")}, *tokenizer.NewPos(9, 11)),
		},
		{
			input:   "1 my_field map(string,bool)",
			setLine: 5,
			inputAnnotationOrStatements: map[int]*[]annotationOrComment{
				0: {{comment: &CommentStmt{*token.NewToken(token.Comment, "this is not documentation"), *tokenizer.NewPos()}}},
				3: {{comment: &CommentStmt{*token.NewToken(token.Comment, "this is documentation"), *tokenizer.NewPos(0, 0, 3, 3)}}},
				4: {
					{comment: &CommentStmt{*token.NewToken(token.Comment, "this is documentation too"), *tokenizer.NewPos(0, 0, 4, 4)}},
					{comment: &CommentStmt{*token.NewToken(token.Comment, "hello"), *tokenizer.NewPos(0, 0, 4, 4)}},
				},
			},
			want: &FieldStmt{
				Index: &IdentStmt{*token.NewToken(token.Integer, "1"), *tokenizer.NewPos(0, 1, 5, 5)},
				Name:  IdentStmt{*token.NewToken(token.Ident, "my_field"), *tokenizer.NewPos(2, 10)},
				ValueType: &DeclStmt{*token.NewToken(token.Ident, "map"), *tokenizer.NewPos(11, 27), []DeclStmt{
					{*token.NewToken(token.Ident, "string"), *tokenizer.NewPos(15, 21), nil, nil, false},
					{*token.NewToken(token.Ident, "bool"), *tokenizer.NewPos(22, 26), nil, nil, false},
				}, nil, false},
				Documentation: []CommentStmt{
					{*token.NewToken(token.Comment, "this is documentation"), *tokenizer.NewPos(0, 0, 3, 3)},
					{*token.NewToken(token.Comment, "hello"), *tokenizer.NewPos(0, 0, 4, 4)},
					{*token.NewToken(token.Comment, "this is documentation too"), *tokenizer.NewPos(0, 0, 4, 4)},
				},
				Annotations: nil,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := newParser(tt.input)
			parser.Reset()

			if tt.inputAnnotationOrStatements != nil {
				for k, v := range tt.inputAnnotationOrStatements {
					parser.annotationsOrComments.Set(k, v)
				}
			}

			parser.currentToken.position.Line = tt.setLine
			parser.currentToken.position.Endline = tt.setLine

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

func TestParser_ParseDefaultsBlock(t *testing.T) {
	tests := []struct {
		input   string
		want    []AssignStmt
		wantErr *ParserError
	}{
		{
			input: `defaults {my_field = "hello" another = {"first": true, "last": 2.21} list = ["a", true]}`,
			want: []AssignStmt{
				{
					Token: *token.NewToken(token.Assign),
					Pos:   *tokenizer.NewPos(10, 28),
					Left:  IdentStmt{*token.NewToken(token.Ident, "my_field"), *tokenizer.NewPos(10, 18)},
					Right: LiteralStmt{*token.NewToken(token.String, "hello"), StringLiteral{"hello"}, *tokenizer.NewPos(21, 28)},
				},
				{
					Token: *token.NewToken(token.Assign),
					Pos:   *tokenizer.NewPos(29, 68),
					Left:  IdentStmt{*token.NewToken(token.Ident, "another"), *tokenizer.NewPos(29, 36)},
					Right: LiteralStmt{
						Token: *token.NewToken(token.Map),
						Pos:   *tokenizer.NewPos(39, 68),
						Kind: MapLiteral{
							{
								Key:   LiteralStmt{*token.NewToken(token.String, "first"), StringLiteral{"first"}, *tokenizer.NewPos(40, 47)},
								Value: LiteralStmt{*token.NewToken(token.Ident, "true"), BooleanLiteral{true}, *tokenizer.NewPos(49, 53)},
							},
							{
								Key:   LiteralStmt{*token.NewToken(token.String, "last"), StringLiteral{"last"}, *tokenizer.NewPos(55, 61)},
								Value: LiteralStmt{*token.NewToken(token.Decimal, "2.21"), FloatLiteral{2.21}, *tokenizer.NewPos(63, 67)},
							},
						},
					},
				},
				{
					Token: *token.NewToken(token.Assign),
					Pos:   *tokenizer.NewPos(69, 87),
					Left: IdentStmt{
						Token: *token.NewToken(token.Ident, "list"),
						Pos:   *tokenizer.NewPos(69, 73),
					},
					Right: LiteralStmt{
						Token: *token.NewToken(token.List),
						Pos:   *tokenizer.NewPos(76, 87),
						Kind: ListLiteral{
							{*token.NewToken(token.String, "a"), StringLiteral{"a"}, *tokenizer.NewPos(77, 80)},
							{*token.NewToken(token.Ident, "true"), BooleanLiteral{true}, *tokenizer.NewPos(82, 86)},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := newParser(tt.input)
			parser.Reset()
			parser.next()

			stmts := parser.parseDefaultsBlock()
			if tt.wantErr == nil {
				require.Empty(t, parser.errors)
			} else {
				require.NotEmpty(t, parser.errors)
				require.Equal(t, *tt.wantErr, *(*parser.errors)[0])
			}

			if diff := cmp.Diff(tt.want, stmts, literalKindExporter); diff != "" {
				t.Errorf("TestParser_ParseDefaultsBlock: %s -> mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestParser_ParseTypeStmt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *TypeStmt
		wantErr *ParserError
	}{
		{
			input: "type MyStruct struct {my_name string}",
			want: &TypeStmt{
				Name:          IdentStmt{*token.NewToken(token.Ident, "MyStruct"), *tokenizer.NewPos(5, 13)},
				Modifier:      token.Struct,
				BaseType:      nil,
				Documentation: nil,
				Annotations:   nil,
				Defaults:      nil,
				Fields: []FieldStmt{
					{
						Index: nil,
						Name:  IdentStmt{*token.NewToken(token.Ident, "my_name"), *tokenizer.NewPos(22, 29)},
						ValueType: &DeclStmt{
							Token:    *token.NewToken(token.Ident, "string"),
							Pos:      *tokenizer.NewPos(30, 36),
							Args:     nil,
							Alias:    nil,
							Nullable: false,
						},
					},
				},
			},
		},
		{
			input:   "type MyStruct struct {my_name string",
			want:    nil,
			wantErr: NewParserErr(ErrUnexpectedEOF{}, *tokenizer.NewPos(36, 36)),
		},
		{
			input: "type MyStruct struct {}",
			want: &TypeStmt{
				Name:          IdentStmt{*token.NewToken(token.Ident, "MyStruct"), *tokenizer.NewPos(5, 13)},
				Modifier:      token.Struct,
				BaseType:      nil,
				Documentation: nil,
				Annotations:   nil,
				Defaults:      nil,
				Fields:        nil,
			},
		},
		{
			name: "base type",
			input: `
			// this is my base type for any other entity
            type Base base {
                // The id of the entity
                id string

                #obsolete = true
                modified_at Time

                //A comment and a:
                //another one
                #hello = 21
                both bool
            }`,
			want: &TypeStmt{
				Name:     IdentStmt{*token.NewToken(token.Ident, "Base"), *tokenizer.NewPos()},
				Modifier: token.Base,
				BaseType: nil,
				Documentation: []CommentStmt{
					{Token: *token.NewToken(token.Comment, " this is my base type for any other entity")},
				},
				Annotations: nil,
				Defaults:    nil,
				Fields: []FieldStmt{
					{
						Name:      IdentStmt{Token: *token.NewToken(token.Ident, "id")},
						ValueType: &DeclStmt{Token: *token.NewToken(token.Ident, "string")},
						Documentation: []CommentStmt{
							{Token: *token.NewToken(token.Comment, " The id of the entity")},
						},
					},
					{
						Name:          IdentStmt{Token: *token.NewToken(token.Ident, "modified_at")},
						ValueType:     &DeclStmt{Token: *token.NewToken(token.Ident, "Time")},
						Documentation: nil,
						Annotations: []AnnotationStmt{
							{
								Token: *token.NewToken(token.Hash),
								Assigment: AssignStmt{
									Token: *token.NewToken(token.Assign),
									Left:  IdentStmt{Token: *token.NewToken(token.Ident, "obsolete")},
									Right: LiteralStmt{
										Token: *token.NewToken(token.Ident, "true"),
										Kind:  BooleanLiteral{true},
									},
								},
							},
						},
					},
					{
						Name:      IdentStmt{Token: *token.NewToken(token.Ident, "both")},
						ValueType: &DeclStmt{Token: *token.NewToken(token.Ident, "bool")},
						Documentation: []CommentStmt{
							{Token: *token.NewToken(token.Comment, "A comment and a:")},
							{Token: *token.NewToken(token.Comment, "another one")},
						},
						Annotations: []AnnotationStmt{
							{
								Token: *token.NewToken(token.Hash),
								Assigment: AssignStmt{
									Token: *token.NewToken(token.Assign),
									Left:  IdentStmt{Token: *token.NewToken(token.Ident, "hello")},
									Right: LiteralStmt{
										Token: *token.NewToken(token.Integer, "21"),
										Kind:  IntLiteral{21},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "struct extends",
			input: `type User extends Base {id string}`,
			want: &TypeStmt{
				Name:          IdentStmt{*token.NewToken(token.Ident, "User"), *tokenizer.NewPos()},
				Modifier:      token.Struct,
				BaseType:      &DeclStmt{Token: *token.NewToken(token.Ident, "Base")},
				Documentation: nil,
				Annotations:   nil,
				Defaults:      nil,
				Fields: []FieldStmt{
					{
						Name:      IdentStmt{Token: *token.NewToken(token.Ident, "id")},
						ValueType: &DeclStmt{Token: *token.NewToken(token.Ident, "string")},
					},
				},
			},
		},
		{
			name: "missing brace",
			input: `type User extends Base {
				1 id string
				3 both bool

				defaults {
					id = "1234"
					both = true
				
			}`,
			want:    nil,
			wantErr: NewParserErr(ErrUnexpectedEOF{}, *tokenizer.NewPos(4, 4, 8, 8)),
		},
		{
			name: "extends with alias and defaults",
			input: `type User extends foo.Base {
				1 id string

				defaults {
					id = "1234"
				}
				
			}`,
			want: &TypeStmt{
				Name:          IdentStmt{*token.NewToken(token.Ident, "User"), *tokenizer.NewPos()},
				Modifier:      token.Struct,
				BaseType:      &DeclStmt{Token: *token.NewToken(token.Ident, "Base"), Alias: &IdentStmt{Token: *token.NewToken(token.Ident, "foo")}},
				Documentation: nil,
				Annotations:   nil,
				Defaults: []AssignStmt{
					{
						Token: *token.NewToken(token.Assign),
						Left:  IdentStmt{Token: *token.NewToken(token.Ident, "id")},
						Right: LiteralStmt{
							Token: *token.NewToken(token.String, "1234"),
							Kind:  StringLiteral{"1234"},
						},
					},
				},
				Fields: []FieldStmt{
					{
						Index:     &IdentStmt{Token: *token.NewToken(token.Integer, "1")},
						Name:      IdentStmt{Token: *token.NewToken(token.Ident, "id")},
						ValueType: &DeclStmt{Token: *token.NewToken(token.Ident, "string")},
					},
				},
			},
		},
		{
			name: "enum",
			input: `type Color enum {
				unknown
				red
				5 green
			}`,
			want: &TypeStmt{
				Name:          IdentStmt{*token.NewToken(token.Ident, "Color"), *tokenizer.NewPos()},
				Modifier:      token.Enum,
				BaseType:      nil,
				Documentation: nil,
				Annotations:   nil,
				Defaults:      nil,
				Fields: []FieldStmt{
					{Name: IdentStmt{Token: *token.NewToken(token.Ident, "unknown")}},
					{Name: IdentStmt{Token: *token.NewToken(token.Ident, "red")}},
					{
						Index: &IdentStmt{Token: *token.NewToken(token.Integer, "5")},
						Name:  IdentStmt{Token: *token.NewToken(token.Ident, "green")},
					},
				},
			},
		},
		{
			name: "modifier with extends is syntax error",
			input: `type Color enum extends Base {
				unknown
				red
				5 green
			}`,
			want:    nil,
			wantErr: NewParserErr(ErrUnexpectedToken{token.Lbrace, *token.NewToken(token.Extends)}, *tokenizer.NewPos(16, 23)),
		},
	}

	for _, tt := range tests {
		name := tt.name
		if len(name) == 0 {
			name = tt.input
		}
		t.Run(name, func(t *testing.T) {
			parser := newParser(tt.input)
			parser.next()
			parser.next()

			stmts := parser.parseTypeStmt()
			if tt.wantErr == nil {
				require.Empty(t, parser.errors)
			} else {
				require.NotEmpty(t, parser.errors)
				require.Equal(t, *tt.wantErr, *(*parser.errors)[0])
			}

			if diff := cmp.Diff(tt.want, stmts, literalKindExporter, cmp.FilterPath(func(p cmp.Path) bool {
				return strings.Contains(p.String(), "Pos")
			}, cmp.Ignore())); diff != "" {
				t.Errorf("TestParser_ParseTypeStmt: %s -> mismatch (-want +got):\n%s", tt.input, diff)
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
