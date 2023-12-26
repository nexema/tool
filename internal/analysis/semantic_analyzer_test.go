package analysis

import (
	"testing"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
	"tomasweigenast.com/nexema/tool/internal/utils"
)

func TestSemanticAnalyzer_Analyze(t *testing.T) {
	tree := parser.NewParseTree()
	tree.Insert("root", &parser.Ast{
		File: reference.File{Path: "root/a.nex"},
		Statements: []parser.Statement{
			utils.MakeIncludeStatement("root/b.nex", ""),
			utils.MakeAnnotationStatement("obsolete", true),
			utils.MakeTypeStatement("A", token.Struct, "", []parser.Statement{
				utils.MakeCommentStatement("Represents the name of a field"),
				utils.MakeFieldStatement("field_name", 0, "string"),
				utils.MakeFieldStatement("from_other", 1, "B"),
			}),
		},
	})
	tree.Insert("root", &parser.Ast{
		File: reference.File{Path: "root/b.nex"},
		Statements: []parser.Statement{
			utils.MakeTypeStatement("B", token.Enum, "", []parser.Statement{
				utils.MakeFieldStatement("unknown", 0, ""),
				utils.MakeFieldStatement("first", 1, ""),
			}),
		},
	})

	analyzer := NewSemanticAnalyzer(tree)
	analyzer.Analyze()
}
