package analysis

import (
	"fmt"
	"strings"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
)

// SemanticAnalyzer analyzes semantically the parsed ast.
//
// The analysis steps are:
// 1. Include statements are the first statements to be declared
// 2. Types are not declared inside types
// 3. Only one DefaultsStatement per type
// 4. Type names and field names are not duplicated
type SemanticAnalyzer struct {
	rootFolder   string                              // the root folder where the analysis takes part. This is used to resolve import paths relatives to this
	tree         *parser.ParseTree                   // the original ParseTree
	symbolTable  *symbolTable                        // the ast's symbol table
	dependencies *dependencyGraph                    // the dependency graph of the whole ParseTree
	contexts     map[reference.File]*analyzerContext // the list of created analyzerContext
}

// NewSemanticAnalyzer creates a new SemanticAnalyzer for the given tree resolving
// for the given rootFolder.
func NewSemanticAnalyzer(tree *parser.ParseTree, rootFolder string) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		rootFolder:   rootFolder,
		tree:         tree,
		symbolTable:  &symbolTable{},
		dependencies: &dependencyGraph{},
		contexts:     make(map[reference.File]*analyzerContext),
	}
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

	// check for cyclic dependencies
	self.checkCyclicDependencies()

	// resolve unresolved references
	self.resolveReferences()
}

// scanSymbols scan top level Type statements of every Ast file and pushes them to the global symbolTable
func (self *SemanticAnalyzer) scanSymbols(pkgName string, node *parser.ParseNode) {
	for _, ast := range node.AstList {
		for _, statement := range ast.Statements {
			if typeStatement, ok := statement.(*parser.TypeStatement); ok {
				self.symbolTable.push(ast.File, selfSymbols, newTypeSymbol(typeStatement))
			}
		}
	}

	node.Iter(func(pkgName string, node *parser.ParseNode) {
		self.scanSymbols(pkgName, node)
	})
}

func (self *SemanticAnalyzer) analyzeNode(pkgName string, node *parser.ParseNode) {
	for _, ast := range node.AstList {
		context := &analyzerContext{
			parent:               self,
			ast:                  ast,
			commentsForNext:      make([]*parser.CommentStatement, 0),
			annotationsForNext:   make([]*parser.AnnotationStatement, 0),
			resolvedImports:      make([]include, 0),
			aliases:              make(map[string]bool),
			unresolvedReferences: make([]unresolvedReference, 0),
		}
		self.contexts[context.ast.File] = context
		context.analyze()
	}

	node.Iter(func(pkgName string, node *parser.ParseNode) {
		self.analyzeNode(pkgName, node)
	})
}

func (self *SemanticAnalyzer) checkCyclicDependencies() {
	cyclicDependencies := self.dependencies.findCyclicDependencies()
	if len(cyclicDependencies) > 0 {
		fmt.Println("Cyclic dependencies detected:")
		for _, cycle := range cyclicDependencies {
			fmt.Printf("%s\n", strings.Join(cycle, " -> "))
		}

		panic("")
	}
}

func (self *SemanticAnalyzer) resolveReferences() {
	for _, analyzerContext := range self.contexts {
		// fill symbol table
		for _, include := range analyzerContext.resolvedImports {
			// get node symbols
			symbols := self.symbolTable.getSelfSymbols(include.path)

			// push to "this" symbolTable
			self.symbolTable.pushAll(analyzerContext.ast.File, include.alias, symbols)
		}

		// resolve references
		for _, ref := range analyzerContext.unresolvedReferences {
			// lookup type
			aliases := self.symbolTable.of(analyzerContext.ast.File)
			if aliases == nil {
				panic("symbolTable not found")
			}

			symbol := aliases.lookup(ref.typeName, ref.alias)
			if symbol == nil {
				panic(fmt.Errorf("symbol %q not found", ref.typeName))
			}
		}
	}
}
