package rules

import (
	analyzer_error "tomasweigenast.com/nexema/tool/internal/analyzer/error"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// BaseAnalyzerRule is an interface that must be implemented for each validation rule
// T is the value that must be passed for validation
type BaseAnalyzerRule interface {
	Validate(value any) []analyzer_error.AnalyzerErrorKind
}

type ObjectResolver func(decl *parser.DeclStmt) *scope.Object

// AnalyzerRuleCollection groups rules which validates over the same value T
type AnalyzerRuleCollection struct {
	ResolveObject ObjectResolver
	Rules         []BaseAnalyzerRule
}
