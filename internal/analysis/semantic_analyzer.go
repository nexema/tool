package analysis

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/token"
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

// analyzerContext groups related statements and analyzes them. In this case, an Ast
type analyzerContext struct {
	parent          *SemanticAnalyzer
	ast             *parser.Ast      // the ast that is being analyzed
	parentStatement parser.Statement // the parent statement of the current statement that is being analyzed

	includeStatementsRead bool                          // a flag that indicates if include statements were read
	defaultsStatementRead bool                          // a flag that indicates if the defaults statement were read
	commentsForNext       []*parser.CommentStatement    // the list of comments (single line) that will be added to the next object
	annotationsForNext    []*parser.AnnotationStatement // the list of annotations that will be added to the next object
}

// NewSemanticAnalyzer creates a new SemanticAnalyzer for the given tree
func NewSemanticAnalyzer(tree *parser.ParseTree) *SemanticAnalyzer {
	return &SemanticAnalyzer{tree}
}

func newAnalyzerContext(semanticAnalyzer *SemanticAnalyzer, ast *parser.Ast) *analyzerContext {
	return &analyzerContext{
		parent:             semanticAnalyzer,
		ast:                ast,
		commentsForNext:    make([]*parser.CommentStatement, 0),
		annotationsForNext: make([]*parser.AnnotationStatement, 0),
	}
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
		context := newAnalyzerContext(self, ast)
		context.analyze()
	}

	node.Iter(func(pkgName string, node *parser.ParseNode) {
		self.analyzeNode(pkgName, node)
	})
}

func (ac *analyzerContext) analyze() {
	ac.annotationsForNext = make([]*parser.AnnotationStatement, 0)
	ac.commentsForNext = make([]*parser.CommentStatement, 0)

	// iterate over every statement declared in the Ast
	for _, statement := range ac.ast.Statements {
		ac.analyzeStatement(statement)
	}
}

func (ac *analyzerContext) analyzeStatement(statement parser.Statement) {
	switch kind := statement.(type) {
	case *parser.IncludeStatement:
		if ac.includeStatementsRead {
			panic("include statements read") // todo: change to err handle
		}

	case *parser.TypeStatement: // todo: add ServiceStatement when added
		ac.includeStatementsRead = true
		if ac.parentStatement != nil {
			panic("type inside type!")
		}

		ac.parentStatement = kind
		for _, body := range kind.Body.Statements {
			ac.analyzeStatement(body)
		}

		ac.parentStatement = nil

	case *parser.AnnotationStatement, *parser.CommentStatement:
		ac.includeStatementsRead = true
		if comment, ok := kind.(*parser.CommentStatement); ok && comment.Token.Kind == token.Comment {
			ac.commentsForNext = append(ac.commentsForNext, comment)
		}

		if annotation, ok := kind.(*parser.AnnotationStatement); ok {
			ac.annotationsForNext = append(ac.annotationsForNext, annotation)
		}

	case *parser.FieldStatement:
		ac.includeStatementsRead = true
		parentTypeStatement, ok := ac.parentStatement.(*parser.TypeStatement)
		if !ok {
			panic("not in a type!")
		}

		_ = parentTypeStatement

	case *parser.DefaultsStatement:
		ac.includeStatementsRead = true
		parentTypeStatement, ok := ac.parentStatement.(*parser.TypeStatement)
		if !ok {
			panic("not in a type!")
		}

		if ac.defaultsStatementRead {
			panic("defaults duplicated")
		}

		ac.defaultsStatementRead = true

		_ = parentTypeStatement
	}
}
