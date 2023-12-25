package parser

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
)

var literalKindExporter = cmp.AllowUnexported(BooleanLiteral{}, StringLiteral{}, IntLiteral{}, FloatLiteral{})

func TestParser_Parse(t *testing.T) {
	input := `
	include "path"
	include "path" as foo
	include "path/faa" as foo_bar
	
	#obsolete = true
	type A {
		// Represents the name of a field
		0 field_name string
	}

	type B enum {
		unknown
		first
	
		#replacement = "first"
		two
	}
	`
	want := &Ast{
		File: reference.File{Path: ":test:"},
		Statements: []Statement{
			&IncludeStatement{
				Token: *token.NewToken(token.Include),
				Path:  LiteralStatement{Token: *token.NewToken(token.String, "path"), Value: StringLiteral{"path"}},
			},
			&IncludeStatement{
				Token: *token.NewToken(token.Include),
				Alias: &IdentifierStatement{Token: *token.NewToken(token.Ident, "foo")},
				Path:  LiteralStatement{Token: *token.NewToken(token.String, "path"), Value: StringLiteral{"path"}},
			},
			&IncludeStatement{
				Token: *token.NewToken(token.Include),
				Alias: &IdentifierStatement{Token: *token.NewToken(token.Ident, "foo_bar")},
				Path:  LiteralStatement{Token: *token.NewToken(token.String, "path/faa"), Value: StringLiteral{"path/faa"}},
			},
			&AnnotationStatement{
				Token: *token.NewToken(token.Hash),
				Assignation: &AssignStatement{
					Token:      *token.NewToken(token.Assign),
					Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "obsolete")},
					Value:      &LiteralStatement{Token: *token.NewToken(token.Ident, "true"), Value: BooleanLiteral{true}},
				},
			},
			&TypeStatement{
				Token: *token.NewToken(token.Type),
				Name:  IdentifierStatement{Token: *token.NewToken(token.Ident, "A")},
				Body: &BlockStatement{
					Token: *token.NewToken(token.Lbrace),
					Statements: []Statement{
						&CommentStatement{Token: *token.NewToken(token.Comment, " Represents the name of a field")},
						&FieldStatement{
							Token: *token.NewToken(token.Ident, "field_name"),
							Index: &LiteralStatement{Token: *token.NewToken(token.Integer, "0"), Value: IntLiteral{int64(0)}},
							ValueType: &DeclarationStatement{
								Token:      *token.NewToken(token.Ident, "string"),
								Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
							},
						},
					},
				},
			},
			&TypeStatement{
				Token:    *token.NewToken(token.Type),
				Name:     IdentifierStatement{Token: *token.NewToken(token.Ident, "B")},
				Modifier: &IdentifierStatement{Token: *token.NewToken(token.Enum)},
				Body: &BlockStatement{
					Token: *token.NewToken(token.Lbrace),
					Statements: []Statement{
						&FieldStatement{Token: *token.NewToken(token.Ident, "unknown")},
						&FieldStatement{Token: *token.NewToken(token.Ident, "first")},
						&AnnotationStatement{
							Token: *token.NewToken(token.Hash),
							Assignation: &AssignStatement{
								Token:      *token.NewToken(token.Assign),
								Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "replacement")},
								Value:      &LiteralStatement{Token: *token.NewToken(token.String, "first"), Value: StringLiteral{"first"}},
							},
						},
						&FieldStatement{Token: *token.NewToken(token.Ident, "two")},
					},
				},
			},
		},
	}

	parser := newParser(input)
	ast := parser.Parse()
	checkParserErrors(t, parser)
	if ast == nil {
		t.Fatalf("Parse() returned nil")
	}

	if diff := cmp.Diff(want, ast, literalKindExporter); diff != "" {
		t.Errorf("Parse(%s) mismatch (-want +got):\n%s", input, diff)
	}

}

func TestParser_ParseInclude(t *testing.T) {
	input := `
	include "path"
	include "path" as foo
	include "path/faa" as foo_bar
	`
	parser := newParser(input)
	ast := parser.Parse()
	checkParserErrors(t, parser)
	if ast == nil {
		t.Fatalf("Parse() returned nil")
	}

	if len(ast.Statements) != 3 {
		t.Fatalf("expected len(ast.Statements) to be 3, got %d", len(ast.Statements))
	}

	tests := []struct {
		expectPath  string
		expectAlias string
	}{
		{"path", ""},
		{"path", "foo"},
		{"path/faa", "foo_bar"},
	}

	for i, test := range tests {
		statement := ast.Statements[i]
		include, ok := statement.(*IncludeStatement)
		if !ok {
			t.Fatalf("expected IncludeStatement, got %s", statement.TokenLiteral())
		}

		path, ok := include.Path.Value.Value().(string)
		if !ok {
			t.Fatalf("expected path to be string, got %d", include.Path.Value)
		}

		if path != test.expectPath {
			t.Fatalf("expected path to be %s, got %s", test.expectPath, path)

		}

		if len(test.expectAlias) == 0 && include.Alias != nil {
			t.Fatalf("expected no alias, got %s", include.Alias.Token.Literal)
		}

		if len(test.expectAlias) != 0 {
			if include.Alias == nil {
				t.Fatalf("expected to have an alias, but none was found")
			}

			if test.expectAlias != include.Alias.Token.Literal {
				t.Fatalf("expected alias to be %s, got %s", test.expectAlias, include.Alias.Token.Literal)
			}
		}
	}
}

func TestParser_ParseCommentStatement(t *testing.T) {
	tests := []struct {
		input  string
		expect *CommentStatement
	}{
		{
			input: `// hello world`,
			expect: &CommentStatement{
				Token: *token.NewToken(token.Comment, " hello world"),
			},
		},
		{
			input: `// hello // world`,
			expect: &CommentStatement{
				Token: *token.NewToken(token.Comment, " hello // world"),
			},
		},
		{
			input: `/* multiline single */`,
			expect: &CommentStatement{
				Token: *token.NewToken(token.CommentMultiline, " multiline single "),
			},
		},
		{
			input: `/* multiline
comment*/`,
			expect: &CommentStatement{
				Token: *token.NewToken(token.CommentMultiline, " multiline\ncomment"),
			},
		},
		{
			input: `/* multiline
comment // inside lines /* other
*/`,
			expect: &CommentStatement{
				Token: *token.NewToken(token.CommentMultiline, " multiline\ncomment // inside lines /* other\n"),
			},
		},
	}

	for _, test := range tests {
		parser := newParser(test.input)
		statement := parser.parseCommentStatement()
		checkParserErrors(t, parser)
		if diff := cmp.Diff(test.expect, statement, literalKindExporter); diff != "" {
			t.Errorf("parseCommentStatement(%s) mismatch (-want +got):\n%s", test.input, diff)
		}
	}
}

func TestParser_ParseAnnotationStatement(t *testing.T) {

	tests := []struct {
		input  string
		expect *AnnotationStatement
	}{
		{
			input: `#key = "hello"`,
			expect: &AnnotationStatement{
				Token: *token.NewToken(token.Hash),
				Assignation: &AssignStatement{
					Token:      *token.NewToken(token.Assign),
					Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "key")},
					Value:      &LiteralStatement{Token: *token.NewToken(token.String, "hello"), Value: StringLiteral{"hello"}},
				},
			},
		},
		{
			input: `#key = 25.4`,
			expect: &AnnotationStatement{
				Token: *token.NewToken(token.Hash),
				Assignation: &AssignStatement{
					Token:      *token.NewToken(token.Assign),
					Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "key")},
					Value:      &LiteralStatement{Token: *token.NewToken(token.Decimal, "25.4"), Value: FloatLiteral{float64(25.4)}},
				},
			},
		},
		{
			input: `#key = [.3, true, "hi"]`,
			expect: &AnnotationStatement{
				Token: *token.NewToken(token.Hash),
				Assignation: &AssignStatement{
					Token:      *token.NewToken(token.Assign),
					Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "key")},
					Value: &LiteralStatement{Token: *token.NewToken(token.Lbrack), Value: ListLiteral{
						FloatLiteral{float64(.3)}, BooleanLiteral{true}, StringLiteral{"hi"},
					}},
				},
			},
		},
	}

	for _, test := range tests {
		parser := newParser(test.input)
		statement := parser.parseAnnotationStatement()
		checkParserErrors(t, parser)
		if diff := cmp.Diff(test.expect, statement, literalKindExporter); diff != "" {
			t.Errorf("parseAnnotationStatement(%s) mismatch (-want +got):\n%s", test.input, diff)
		}
	}
}

func TestParser_ParseTypeStatement(t *testing.T) {

	tests := []struct {
		input  string
		expect *TypeStatement
	}{
		{
			input: `type A {}`,
			expect: &TypeStatement{
				Token: *token.NewToken(token.Type),
				Name:  IdentifierStatement{Token: *token.NewToken(token.Ident, "A")},
				Body:  emptyBlockStatement(),
			},
		},
		{
			input: `type A struct {}`,
			expect: &TypeStatement{
				Token:    *token.NewToken(token.Type),
				Name:     IdentifierStatement{Token: *token.NewToken(token.Ident, "A")},
				Modifier: &IdentifierStatement{Token: *token.NewToken(token.Struct)},
				Body:     emptyBlockStatement(),
			},
		},
		{
			input: `type A union {}`,
			expect: &TypeStatement{
				Token:    *token.NewToken(token.Type),
				Name:     IdentifierStatement{Token: *token.NewToken(token.Ident, "A")},
				Modifier: &IdentifierStatement{Token: *token.NewToken(token.Union)},
				Body:     emptyBlockStatement(),
			},
		},
		{
			input: `type A enum {}`,
			expect: &TypeStatement{
				Token:    *token.NewToken(token.Type),
				Name:     IdentifierStatement{Token: *token.NewToken(token.Ident, "A")},
				Modifier: &IdentifierStatement{Token: *token.NewToken(token.Enum)},
				Body:     emptyBlockStatement(),
			},
		},
		{
			input: `type A base {}`,
			expect: &TypeStatement{
				Token:    *token.NewToken(token.Type),
				Name:     IdentifierStatement{Token: *token.NewToken(token.Ident, "A")},
				Modifier: &IdentifierStatement{Token: *token.NewToken(token.Base)},
				Body:     emptyBlockStatement(),
			},
		},
		{
			input: `type A extends Foo {}`,
			expect: &TypeStatement{
				Token:   *token.NewToken(token.Type),
				Name:    IdentifierStatement{Token: *token.NewToken(token.Ident, "A")},
				Extends: &ExtendsStatement{Token: *token.NewToken(token.Extends), BaseType: IdentifierStatement{Token: *token.NewToken(token.Ident, "Foo")}},
				Body:    emptyBlockStatement(),
			},
		},
		{
			input: `type A extends foo_bar.Foo {}`,
			expect: &TypeStatement{
				Token:   *token.NewToken(token.Type),
				Name:    IdentifierStatement{Token: *token.NewToken(token.Ident, "A")},
				Extends: &ExtendsStatement{Token: *token.NewToken(token.Extends), BaseType: IdentifierStatement{Token: *token.NewToken(token.Ident, "Foo"), Alias: token.NewToken(token.Ident, "foo_bar")}},
				Body:    emptyBlockStatement(),
			},
		},
		{
			input: `type Fields {
				0 field_name string
				field_name int64?
			}`,
			expect: &TypeStatement{
				Token: *token.NewToken(token.Type),
				Name:  IdentifierStatement{Token: *token.NewToken(token.Ident, "Fields")},
				Body: &BlockStatement{
					Token: *token.NewToken(token.Lbrace),
					Statements: []Statement{
						&FieldStatement{
							Token:     *token.NewToken(token.Ident, "field_name"),
							Index:     &LiteralStatement{Token: *token.NewToken(token.Integer, "0"), Value: IntLiteral{int64(0)}},
							ValueType: &DeclarationStatement{Token: *token.NewToken(token.Ident, "string"), Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")}},
						},
						&FieldStatement{
							Token:     *token.NewToken(token.Ident, "field_name"),
							ValueType: &DeclarationStatement{Token: *token.NewToken(token.Ident, "int64"), Nullable: true, Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "int64")}},
						},
					},
				},
			},
		},
		{
			input: `type FieldsDefaults {
				field_name string
				defaults {
					"field_name": "hello",
					"age": 12,
					"nested": {
						12: true,
						4: false,
						1: ["a", "b", 24.32]
					}
				}
			}`,
			expect: &TypeStatement{
				Token: *token.NewToken(token.Type),
				Name:  IdentifierStatement{Token: *token.NewToken(token.Ident, "FieldsDefaults")},
				Body: &BlockStatement{
					Token: *token.NewToken(token.Lbrace),
					Statements: []Statement{
						&FieldStatement{
							Token:     *token.NewToken(token.Ident, "field_name"),
							ValueType: &DeclarationStatement{Token: *token.NewToken(token.Ident, "string"), Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")}},
						},
						&DefaultsStatement{
							Token: *token.NewToken(token.Defaults),
							Values: &LiteralStatement{
								Token: *token.NewToken(token.Lbrace),
								Value: MapLiteral{
									{StringLiteral{"field_name"}, StringLiteral{"hello"}},
									{StringLiteral{"age"}, IntLiteral{int64(12)}},
									{StringLiteral{"nested"}, MapLiteral{
										{IntLiteral{int64(12)}, BooleanLiteral{true}},
										{IntLiteral{int64(4)}, BooleanLiteral{false}},
										{IntLiteral{int64(1)}, ListLiteral{
											StringLiteral{"a"}, StringLiteral{"b"}, FloatLiteral{float64(24.32)},
										}},
									}},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		parser := newParser(test.input)
		statement := parser.parseTypeStatement()
		checkParserErrors(t, parser)
		if diff := cmp.Diff(test.expect, statement, literalKindExporter); diff != "" {
			t.Errorf("parseTypeStatement(%s) mismatch (-want +got):\n%s", test.input, diff)
		}
	}
}

func TestParser_ParseFieldStatement(t *testing.T) {
	tests := []struct {
		input  string
		expect *FieldStatement
	}{
		{
			input: "5 field_name string",
			expect: &FieldStatement{
				Token: *token.NewToken(token.Ident, "field_name"),
				Index: &LiteralStatement{Token: *token.NewToken(token.Integer, "5"), Value: IntLiteral{int64(5)}},
				ValueType: &DeclarationStatement{
					Token:      *token.NewToken(token.Ident, "string"),
					Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
				},
			},
		},
		{
			input: "field_name string",
			expect: &FieldStatement{
				Token: *token.NewToken(token.Ident, "field_name"),
				Index: nil,
				ValueType: &DeclarationStatement{
					Token:      *token.NewToken(token.Ident, "string"),
					Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
				},
			},
		},
		{
			input: "field_name list(string)?",
			expect: &FieldStatement{
				Token: *token.NewToken(token.Ident, "field_name"),
				Index: nil,
				ValueType: &DeclarationStatement{
					Token:      *token.NewToken(token.Ident, "list"),
					Nullable:   true,
					Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "list")},
					Arguments: []DeclarationStatement{
						{
							Token:      *token.NewToken(token.Ident, "string"),
							Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
						},
					},
				},
			},
		},
		{
			input: "enum_field_value",
			expect: &FieldStatement{
				Token:     *token.NewToken(token.Ident, "enum_field_value"),
				Index:     nil,
				ValueType: nil,
			},
		},
	}

	for _, test := range tests {
		parser := newParser(test.input)

		statement := parser.parseFieldStatement()
		if statement == nil {
			t.Fatalf("expected FieldStatement, got nil")
		}
		checkParserErrors(t, parser)
		if diff := cmp.Diff(test.expect, statement, literalKindExporter); diff != "" {
			t.Errorf("parseFieldStatement() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestParser_ParseDeclarationStatement(t *testing.T) {
	tests := []struct {
		input  string
		expect *DeclarationStatement
	}{
		{
			input: "string",
			expect: &DeclarationStatement{
				Token:      *token.NewToken(token.Ident, "string"),
				Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
				Nullable:   false,
			},
		},
		{
			input: "string?",
			expect: &DeclarationStatement{
				Token:      *token.NewToken(token.Ident, "string"),
				Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
				Nullable:   true,
			},
		},
		{
			input: "list(string)",
			expect: &DeclarationStatement{
				Token:      *token.NewToken(token.Ident, "list"),
				Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "list")},
				Arguments: []DeclarationStatement{
					{
						Token:      *token.NewToken(token.Ident, "string"),
						Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
					},
				},
			},
		},
		{
			input: "list(string)?",
			expect: &DeclarationStatement{
				Token:      *token.NewToken(token.Ident, "list"),
				Nullable:   true,
				Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "list")},
				Arguments: []DeclarationStatement{
					{
						Token:      *token.NewToken(token.Ident, "string"),
						Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
					},
				},
			},
		},
		{
			input: "list(string?)",
			expect: &DeclarationStatement{
				Token:      *token.NewToken(token.Ident, "list"),
				Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "list")},
				Arguments: []DeclarationStatement{
					{
						Token:      *token.NewToken(token.Ident, "string"),
						Nullable:   true,
						Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
					},
				},
			},
		},
		{
			input: "list(string?)?",
			expect: &DeclarationStatement{
				Token:      *token.NewToken(token.Ident, "list"),
				Nullable:   true,
				Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "list")},
				Arguments: []DeclarationStatement{
					{
						Nullable:   true,
						Token:      *token.NewToken(token.Ident, "string"),
						Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
					},
				},
			},
		},
		{
			input: "map(string, bool?)?",
			expect: &DeclarationStatement{
				Token:      *token.NewToken(token.Ident, "map"),
				Nullable:   true,
				Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "map")},
				Arguments: []DeclarationStatement{
					{
						Token:      *token.NewToken(token.Ident, "string"),
						Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
					},
					{
						Token:      *token.NewToken(token.Ident, "bool"),
						Nullable:   true,
						Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "bool")},
					},
				},
			},
		},
		{
			input: "foo_bar.User",
			expect: &DeclarationStatement{
				Token: *token.NewToken(token.Ident, "User"),
				Identifier: &IdentifierStatement{
					Token: *token.NewToken(token.Ident, "User"),
					Alias: token.NewToken(token.Ident, "foo_bar"),
				},
			},
		},
		{
			input: "foo_bar.User?",
			expect: &DeclarationStatement{
				Token:    *token.NewToken(token.Ident, "User"),
				Nullable: true,
				Identifier: &IdentifierStatement{
					Token: *token.NewToken(token.Ident, "User"),
					Alias: token.NewToken(token.Ident, "foo_bar"),
				},
			},
		},
		{
			input: "map(string, foo_bar.User?)?",
			expect: &DeclarationStatement{
				Token:      *token.NewToken(token.Ident, "map"),
				Nullable:   true,
				Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "map")},
				Arguments: []DeclarationStatement{
					{
						Token:      *token.NewToken(token.Ident, "string"),
						Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
					},
					{
						Token:    *token.NewToken(token.Ident, "User"),
						Nullable: true,
						Identifier: &IdentifierStatement{
							Token: *token.NewToken(token.Ident, "User"),
							Alias: token.NewToken(token.Ident, "foo_bar"),
						},
					},
				},
			},
		},
		{
			input: `varchar(1234)`,
			expect: &DeclarationStatement{
				Token:      *token.NewToken(token.Ident, "varchar"),
				Identifier: &IdentifierStatement{Token: *token.NewToken(token.Ident, "varchar")},
				Arguments: []DeclarationStatement{
					{
						Token:      *token.NewToken(token.Integer, "1234"),
						Identifier: &IdentifierStatement{Token: *token.NewToken(token.Integer, "1234")},
					},
				},
			},
		},
	}

	for _, test := range tests {
		parser := newParser(test.input)

		statement := parser.parseDeclarationStatement()
		if statement == nil {
			t.Fatalf("expected DeclarationStatement, got nil")
		}
		checkParserErrors(t, parser)
		if diff := cmp.Diff(test.expect, statement, literalKindExporter); diff != "" {
			t.Errorf("parseDeclarationStatement() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestParser_ParseLiteralStatement(t *testing.T) {
	tests := []struct {
		input  string
		expect Literal
	}{
		{
			input:  "5",
			expect: IntLiteral{int64(5)},
		},
		{
			input:  "25.4",
			expect: FloatLiteral{float64(25.4)},
		},
		{
			input:  ".4",
			expect: FloatLiteral{float64(.4)},
		},
		{
			input:  "true",
			expect: BooleanLiteral{true},
		},
		{
			input:  "false",
			expect: BooleanLiteral{false},
		},
		{
			input:  `"hello"`,
			expect: StringLiteral{"hello"},
		},
		{
			input: `[25, 32.2, .4, "one", true, false]`,
			expect: ListLiteral{
				IntLiteral{int64(25)},
				FloatLiteral{float64(32.2)},
				FloatLiteral{float64(.4)},
				StringLiteral{"one"},
				BooleanLiteral{true},
				BooleanLiteral{false},
			},
		},
		{
			input: `{"key":"value", "key": 23.2, "key": .2, "key": true, "a": false, 5: 12.2, 2: "hello"}`,
			expect: MapLiteral{
				{Key: StringLiteral{"key"}, Value: StringLiteral{"value"}},
				{Key: StringLiteral{"key"}, Value: FloatLiteral{float64(23.2)}},
				{Key: StringLiteral{"key"}, Value: FloatLiteral{float64(.2)}},
				{Key: StringLiteral{"key"}, Value: BooleanLiteral{true}},
				{Key: StringLiteral{"a"}, Value: BooleanLiteral{false}},
				{Key: IntLiteral{int64(5)}, Value: FloatLiteral{float64(12.2)}},
				{Key: IntLiteral{int64(2)}, Value: StringLiteral{"hello"}},
			},
		},
	}

	for _, test := range tests {
		parser := newParser(test.input)

		statement := parser.parseLiteralStatement()
		if statement == nil {
			t.Fatalf("expected LiteralStatement, got nil")
		}
		checkParserErrors(t, parser)

		if !reflect.DeepEqual(test.expect.Value(), statement.Value.Value()) {
			t.Fatalf("expected literal to be %v [type: %s], but got %v [type: %s]", test.expect.Value(), reflect.TypeOf(test.expect.Value()).Name(), statement.Value.Value(), reflect.TypeOf(statement.Value.Value()).Name())
		}
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.errors
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg.Kind.Message())
	}
	t.FailNow()
}

func newParser(i string) *Parser {
	p := NewParser(bytes.NewBufferString(i), reference.File{
		Path: ":test:",
	})

	return p
}

func newint64(v int64) *int64 {
	return &v
}

func emptyBlockStatement() *BlockStatement {
	return &BlockStatement{
		Token:      *token.NewToken(token.Lbrace),
		Statements: []Statement{},
	}
}
