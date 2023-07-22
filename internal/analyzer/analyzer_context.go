package analyzer

import (
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// AnalyzerContext provides a method to report errors in an analyzer rule, as well provides the parsed scopes to the rule.
type AnalyzerContext struct {
	scope  scope.Scope
	errors *AnalyzerErrorCollection
}

// GetObject under the hood calls FindOjbect on self.currLocalScope and reports any error if any
func (self *AnalyzerContext) GetObject(decl *parser.DeclStmt) *scope.Object {
	name, alias := decl.Format()
	objects := self.scope.FindObject(name, alias)
	if objects == nil {
		self.errors.Push(ErrTypeNotFound{Name: name, Alias: alias}, decl.Pos)
		return nil
	} else if len(objects) > 1 {
		self.errors.Push(ErrNeedAlias{ObjectName: name}, decl.Pos)
		return nil
	} else {
		return objects[0]
	}
}

func (self *AnalyzerContext) RunOver(callback func(object *scope.Object, source *parser.TypeStmt)) {

	for _, obj := range self.scope.GetObjects(1) {
		callback(obj, obj.Source())
	}
}

func (self *AnalyzerContext) RunRule(ruleName string) {

}

func (self *AnalyzerContext) Scope() scope.Scope {
	return self.scope
}

func (self *AnalyzerContext) ReportError(err AnalyzerErrorKind, at reference.Pos) {
	self.errors.Push(err, at)
}

func (self *AnalyzerContext) Errors() *AnalyzerErrorCollection {
	return self.errors
}

// NewAnalyzerContext creates a new AnalyzerContext
func NewAnalyzerContext(scope scope.Scope) *AnalyzerContext {
	return &AnalyzerContext{scope, NewAnalyzerErrorCollection()}
}
