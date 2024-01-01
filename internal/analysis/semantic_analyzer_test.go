package analysis

import (
	"testing"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
	"tomasweigenast.com/nexema/tool/internal/utils"
)

func TestSemanticAnalyzer_Analyze(t *testing.T) {
	// import paths must be relatives to nexema.yaml file, which is considered the root dir
	tree := parser.NewParseTree()
	tree.Insert("v1", &parser.Ast{
		File: reference.File{Path: "v1/a.nex"},
		Statements: []parser.Statement{
			utils.MakeIncludeStatement("v1/b.nex", ""),
			utils.MakeIncludeStatement("v1/c.nex", "foo"),
			utils.MakeAnnotationStatement("obsolete", true),
			utils.MakeTypeStatement("A", token.Struct, "", []parser.Statement{
				utils.MakeCommentStatement("Represents the name of a field"),
				utils.MakeFieldStatement("field_name", 0, "string"),
				utils.MakeFieldStatement("from_other", 1, "B"),
				utils.MakeFieldStatement("from_other_2", 2, "foo.C"),
			}),
		},
	})
	tree.Insert("v1", &parser.Ast{
		File: reference.File{Path: "v1/b.nex"},
		Statements: []parser.Statement{
			utils.MakeTypeStatement("B", token.Enum, "", []parser.Statement{
				utils.MakeFieldStatement("unknown", 0, ""),
				utils.MakeFieldStatement("first", -1, ""),
				utils.MakeFieldStatement("second", 2, ""),
			}),
			utils.MakeTypeStatement("C", token.Struct, "", []parser.Statement{
				utils.MakeFieldStatement("unknown", 0, "string", true),
				utils.MakeFieldStatement("first", 1, "varchar(2048)", true),
				utils.MakeCustomFieldStatement("second", 2, &parser.DeclarationStatement{
					Token:      *token.NewToken(token.Ident, "map"),
					Identifier: &parser.IdentifierStatement{Token: *token.NewToken(token.Ident, "list")},
					Arguments: []parser.DeclarationStatement{
						{
							Token:      *token.NewToken(token.Ident, "string"),
							Identifier: &parser.IdentifierStatement{Token: *token.NewToken(token.Ident, "string")},
						},
						{
							Token:      *token.NewToken(token.Ident, "bool"),
							Identifier: &parser.IdentifierStatement{Token: *token.NewToken(token.Ident, "bool")},
						},
					},
				}),
			}),
		},
	})
	tree.Insert("v1", &parser.Ast{
		File: reference.File{Path: "v1/c.nex"},
		Statements: []parser.Statement{
			utils.MakeTypeStatement("C", token.Enum, "", []parser.Statement{
				utils.MakeFieldStatement("unknown", 0, ""),
				utils.MakeFieldStatement("first", -1, ""),
				utils.MakeFieldStatement("second", 2, ""),
			}),
		},
	})

	analyzer := NewSemanticAnalyzer(tree, "/windows/")
	analyzer.Analyze()
}
