package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
)

var literalKindExporter = cmp.AllowUnexported(BooleanLiteral{}, StringLiteral{}, IntLiteral{}, FloatLiteral{})

func TestParser_Parse(t *testing.T) {
	input := `use "path/to/my/pkg"
	use "foo/bar" as foo_bar

	// An enum with some values
	type MyEnum enum {
		unknown
		first
		second
	}

	// An enum with indexes. This is now obsolete, use MyEnum instead.
	#obsolete=true
	type IndexedEnum enum {
		0 unknown
		1 first
		2 second
	}

	// This is a base type for every struct
	type Base base {
		// The id of the entity
		0 id string

		// The date and time when the type was modified
		1 modified_at Time?
	}

	type User extends Base {
		name string
		tags list(string)
		preferences map(string,bool)?
		calls list(Time?)?

		defaults {
			name = "hello"
			preferences = {
				"animals": false,
				"cars": true
			}
			calls = []
		}
	}

	type User2 union {
		name string
		tags list(string)
	}`

	want := &Ast{
		File: reference.File{
			Path: ":test:",
		},
		UseStatements: []UseStmt{
			{
				Token: *token.NewToken(token.Use),
				Path: LiteralStmt{
					Token: *token.NewToken(token.String, "path/to/my/pkg"),
					Kind:  StringLiteral{"path/to/my/pkg"},
				},
			},
			{
				Token: *token.NewToken(token.Use),
				Path: LiteralStmt{
					Token: *token.NewToken(token.String, "foo/bar"),
					Kind:  StringLiteral{"foo/bar"},
				},
				Alias: &IdentStmt{
					Token: *token.NewToken(token.Ident, "foo_bar"),
				},
			},
		},
		TypeStatements: []TypeStmt{
			{
				Name:     IdentStmt{Token: *token.NewToken(token.Ident, "MyEnum")},
				Modifier: token.Enum,
				BaseType: nil,
				Documentation: []CommentStmt{
					{Token: *token.NewToken(token.Comment, " An enum with some values")},
				},
				Annotations: nil,
				Defaults:    nil,
				Fields: []FieldStmt{
					{Name: IdentStmt{Token: *token.NewToken(token.Ident, "unknown")}},
					{Name: IdentStmt{Token: *token.NewToken(token.Ident, "first")}},
					{Name: IdentStmt{Token: *token.NewToken(token.Ident, "second")}},
				},
			},
			{
				Name:     IdentStmt{Token: *token.NewToken(token.Ident, "IndexedEnum")},
				Modifier: token.Enum,
				BaseType: nil,
				Documentation: []CommentStmt{
					{Token: *token.NewToken(token.Comment, " An enum with indexes. This is now obsolete, use MyEnum instead.")},
				},
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
				Defaults: nil,
				Fields: []FieldStmt{
					{
						Index: &IdentStmt{Token: *token.NewToken(token.Integer, "0")},
						Name:  IdentStmt{Token: *token.NewToken(token.Ident, "unknown")},
					},
					{
						Index: &IdentStmt{Token: *token.NewToken(token.Integer, "1")},
						Name:  IdentStmt{Token: *token.NewToken(token.Ident, "first")},
					},
					{
						Index: &IdentStmt{Token: *token.NewToken(token.Integer, "2")},
						Name:  IdentStmt{Token: *token.NewToken(token.Ident, "second")},
					},
				},
			},
			{
				Name:     IdentStmt{Token: *token.NewToken(token.Ident, "Base")},
				Modifier: token.Base,
				BaseType: nil,
				Documentation: []CommentStmt{
					{Token: *token.NewToken(token.Comment, " This is a base type for every struct")},
				},
				Annotations: nil,
				Defaults:    nil,
				Fields: []FieldStmt{
					{
						Index:     &IdentStmt{Token: *token.NewToken(token.Integer, "0")},
						Name:      IdentStmt{Token: *token.NewToken(token.Ident, "id")},
						ValueType: &DeclStmt{Token: *token.NewToken(token.Ident, "string")},
						Documentation: []CommentStmt{
							{Token: *token.NewToken(token.Comment, " The id of the entity")},
						},
					},
					{
						Index:     &IdentStmt{Token: *token.NewToken(token.Integer, "1")},
						Name:      IdentStmt{Token: *token.NewToken(token.Ident, "modified_at")},
						ValueType: &DeclStmt{Token: *token.NewToken(token.Ident, "Time"), Nullable: true},
						Documentation: []CommentStmt{
							{Token: *token.NewToken(token.Comment, " The date and time when the type was modified")},
						},
					},
				},
			},
			{
				Name:          IdentStmt{Token: *token.NewToken(token.Ident, "User")},
				Modifier:      token.Struct,
				BaseType:      &DeclStmt{Token: *token.NewToken(token.Ident, "Base")},
				Documentation: nil,
				Annotations:   nil,
				Defaults: []AssignStmt{
					{
						Token: *token.NewToken(token.Assign),
						Left:  IdentStmt{Token: *token.NewToken(token.Ident, "name")},
						Right: LiteralStmt{
							Token: *token.NewToken(token.String, "hello"),
							Kind:  StringLiteral{"hello"},
						},
					},
					{
						Token: *token.NewToken(token.Assign),
						Left:  IdentStmt{Token: *token.NewToken(token.Ident, "preferences")},
						Right: LiteralStmt{
							Token: *token.NewToken(token.Map),
							Kind: MapLiteral{
								{
									Key:   LiteralStmt{Token: *token.NewToken(token.String, "animals"), Kind: StringLiteral{"animals"}},
									Value: LiteralStmt{Token: *token.NewToken(token.Ident, "false"), Kind: BooleanLiteral{false}},
								},
								{
									Key:   LiteralStmt{Token: *token.NewToken(token.String, "cars"), Kind: StringLiteral{"cars"}},
									Value: LiteralStmt{Token: *token.NewToken(token.Ident, "true"), Kind: BooleanLiteral{true}},
								},
							},
						},
					},
					{
						Token: *token.NewToken(token.Assign),
						Left:  IdentStmt{Token: *token.NewToken(token.Ident, "calls")},
						Right: LiteralStmt{
							Token: *token.NewToken(token.List),
							Kind:  ListLiteral{},
						},
					},
				},
				Fields: []FieldStmt{
					{
						Name:      IdentStmt{Token: *token.NewToken(token.Ident, "name")},
						ValueType: &DeclStmt{Token: *token.NewToken(token.Ident, "string")},
					},
					{
						Name: IdentStmt{Token: *token.NewToken(token.Ident, "tags")},
						ValueType: &DeclStmt{
							Token: *token.NewToken(token.Ident, "list"),
							Args: []DeclStmt{
								{Token: *token.NewToken(token.Ident, "string")},
							},
						},
					},
					{
						Name: IdentStmt{Token: *token.NewToken(token.Ident, "preferences")},
						ValueType: &DeclStmt{
							Token:    *token.NewToken(token.Ident, "map"),
							Nullable: true,
							Args: []DeclStmt{
								{Token: *token.NewToken(token.Ident, "string")},
								{Token: *token.NewToken(token.Ident, "bool")},
							},
						},
					},
					{
						Name: IdentStmt{Token: *token.NewToken(token.Ident, "calls")},
						ValueType: &DeclStmt{
							Token:    *token.NewToken(token.Ident, "list"),
							Nullable: true,
							Args: []DeclStmt{
								{Token: *token.NewToken(token.Ident, "Time"), Nullable: true},
							},
						},
					},
				},
			},
			{
				Name:          IdentStmt{Token: *token.NewToken(token.Ident, "User2")},
				Modifier:      token.Union,
				BaseType:      nil,
				Documentation: nil,
				Annotations:   nil,
				Defaults:      nil,
				Fields: []FieldStmt{
					{
						Name:      IdentStmt{Token: *token.NewToken(token.Ident, "name")},
						ValueType: &DeclStmt{Token: *token.NewToken(token.Ident, "string")},
					},
					{
						Name: IdentStmt{Token: *token.NewToken(token.Ident, "tags")},
						ValueType: &DeclStmt{
							Token: *token.NewToken(token.Ident, "list"),
							Args: []DeclStmt{
								{Token: *token.NewToken(token.Ident, "string")},
							},
						},
					},
				},
			},
		},
	}

	parser := newParser(input)
	got := parser.Parse()
	require.Empty(t, parser.Errors())
	if diff := cmp.Diff(want, got, literalKindExporter, cmp.FilterPath(func(p cmp.Path) bool {
		return strings.Contains(p.String(), "Pos")
	}, cmp.Ignore())); diff != "" {
		t.Errorf("TestParser_Parse:  mismatch (-want +got):\n%s", diff)
	}
}

func newParser(i string) *Parser {
	p := NewParser(bytes.NewBufferString(i), reference.File{
		Path: ":test:",
	})
	p.Reset()

	return p
}
