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
	tree        *parser.ParseTree // the original ParseTree
	symbolTable *symbolTable      // the ast's symbol table
}

// NewSemanticAnalyzer creates a new SemanticAnalyzer for the given tree
func NewSemanticAnalyzer(tree *parser.ParseTree) *SemanticAnalyzer {
	return &SemanticAnalyzer{tree, &symbolTable{}}
}

// context is the basic interface that every context must implement
//
// A context holds information that helps SemanticAnalyzer when it analyzes a Node
type context interface {
	Statement() parser.Statement // returns the owner Statement of the context
}

func (self *SemanticAnalyzer) Analyze() {
	// TODO: maybe optimize to avoid two iterations over the ParseTree.
	// It can be done by requesting contexts to other nodes of the tree, and if they don't available
	// right now, create at request time. When the Node that uses that context is going to be analyzed,
	// it can use the newly created context.

	// first of all scan for symbols
	self.tree.Iter(func(pkgName string, node *parser.ParseNode) {
		self.scanSymbols(pkgName, node)
	})

	// scan to analyze
	self.tree.Iter(func(pkgName string, node *parser.ParseNode) {
		self.analyzeNode(pkgName, node)
	})
}

func (self *SemanticAnalyzer) scanSymbols(pkgName string, node *parser.ParseNode) {
	fmt.Printf("package: %s\n", pkgName)
	for _, ast := range node.AstList {
		fmt.Printf("\tfile: %s\n", ast.File.Path)
		for _, statement := range ast.Statements {
			if typeStatement, ok := statement.(*parser.TypeStatement); ok {
				self.symbolTable.push(sourceRef{ast.File}, newTypeSymbol(typeStatement))
			}
		}
	}

	node.Iter(func(pkgName string, node *parser.ParseNode) {
		self.scanSymbols(pkgName, node)
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
