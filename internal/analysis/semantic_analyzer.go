package analysis

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/parser"
)

// SemanticAnalyzer analyzes semantically the parsed ast.
//
// The analysis steps are:
// 1. Include statements are the first statements to be declared
// 2. Types are not declared inside types
// 3. Only one DefaultsStatement per type
// 4. Type names and field names are not duplicated
type SemanticAnalyzer struct {
	tree *parser.ParseTree
}

// NewSemanticAnalyzer creates a new SemanticAnalyzer for the given tree
func NewSemanticAnalyzer(tree *parser.ParseTree) *SemanticAnalyzer {
	return &SemanticAnalyzer{tree}
}

// context is the basic interface that every context must implement
//
// A context holds information that helps SemanticAnalyzer when it analyzes a Node
type context interface {
	Statement() parser.Statement // returns the owner Statement of the context
}

func (self *SemanticAnalyzer) Analyze() {
	self.tree.Iter(func(pkgName string, node *parser.ParseNode) {
		self.analyzeNode(pkgName, node)
	})
}

func (self *SemanticAnalyzer) analyzeNode(pkgName string, node *parser.ParseNode) {
	fmt.Printf("package: %s\n", pkgName)
	for _, ast := range node.AstList {
		fmt.Printf("\tfile: %s\n", ast.File.Path)
		context := &analyzerContext{
			parent:             self,
			ast:                ast,
			commentsForNext:    make([]*parser.CommentStatement, 0),
			annotationsForNext: make([]*parser.AnnotationStatement, 0),
			typeNames:          make(map[string]bool),
		}
		context.analyze()
	}

	node.Iter(func(pkgName string, node *parser.ParseNode) {
		self.analyzeNode(pkgName, node)
	})
}
