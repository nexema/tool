package analyzer

import (
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// AnalyzerContext provides a method to report errors in an analyzer rule, as well provides the parsed scopes to the rule.
type AnalyzerContext struct {
	scope  *scope.LocalScope
	errors *AnalyzerErrorCollection
}

// GetObject under the hood calls FindOjbect on self.currLocalScope and reports any error if any
func (self *AnalyzerContext) GetObject(decl *parser.DeclStmt) *scope.Object {
	name, alias := decl.Format()
	obj, needAlias := self.scope.FindObject(name, alias)
	if obj == nil {
		if needAlias {
			self.errors.Push(ErrNeedAlias{}, decl.Pos)
		} else {
			self.errors.Push(ErrTypeNotFound{Name: name, Alias: alias}, decl.Pos)
		}
	} else {
		// todo: may this check if obj is Base and don't allow it to use
		return obj
	}

	return nil
}

func (self *AnalyzerContext) RunOver(callback func(object *scope.Object, source *parser.TypeStmt)) {

	for _, obj := range *self.scope.Objects() {
		src := obj.Source()
		callback(obj, src)
	}
}

func (self *AnalyzerContext) RunRule(ruleName string) {

}

func (self *AnalyzerContext) Scope() *scope.LocalScope {
	return self.scope
}

func (self *AnalyzerContext) ReportError(err AnalyzerErrorKind, at reference.Pos) {
	self.errors.Push(err, at)
}

func (self *AnalyzerContext) Errors() *AnalyzerErrorCollection {
	return self.errors
}

// NewAnalyzerContext creates a new AnalyzerContext
func NewAnalyzerContext(scope *scope.LocalScope) *AnalyzerContext {
	return &AnalyzerContext{scope, NewAnalyzerErrorCollection()}
}
