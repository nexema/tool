package analyzer

import (
	"fmt"
	"path"

	"github.com/mitchellh/hashstructure/v2"
	analyzer_error "tomasweigenast.com/nexema/tool/internal/analyzer/error"
	"tomasweigenast.com/nexema/tool/internal/analyzer/rules"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// Analyzer takes a linked list of built scopes and analyzes them syntactically.
// Also, if analysis succeed, a definition is built
type Analyzer struct {
	scopes []*scope.Scope
	errors *analyzer_error.AnalyzerErrorCollection

	currScope      *scope.Scope
	currLocalScope *scope.LocalScope
	currTypeId     string
	files          []definition.NexemaFile

	rules map[rules.RuleType]*rules.AnalyzerRuleCollection
}

func NewAnalyzer(scopes []*scope.Scope) *Analyzer {
	analyzer := &Analyzer{
		scopes: scopes,
		errors: analyzer_error.NewAnalyzerErrorCollection(),
		files:  make([]definition.NexemaFile, 0),
	}

	analyzer.rules = map[rules.RuleType]*rules.AnalyzerRuleCollection{
		rules.TypeStatementRules: rules.GetTypeStatementRules(analyzer.getObject),
	}

	return analyzer
}

// Analyze starts analyzing and logs any error encountered
func (self *Analyzer) Analyze() {
	for _, scope := range self.scopes {
		self.analyzeScope(scope)
	}
}

func (self *Analyzer) analyzeScope(s *scope.Scope) {
	self.currScope = s
	for _, localScope := range *s.LocalScopes() {
		self.analyzeLocalScope(localScope)
	}
}

func (self *Analyzer) analyzeLocalScope(ls *scope.LocalScope) {
	self.currLocalScope = ls
	file := ls.File()
	nexFile := definition.NexemaFile{
		Types:       make([]definition.TypeDefinition, 0),
		Path:        file.Path,
		PackageName: path.Base(file.Path),
	}

	for _, obj := range *ls.Objects() {
		self.currTypeId = obj.Id

		for _, rule := range self.rules[rules.TypeStatementRules].Rules {
			rule.Validate(obj.Source())
		}

		// rule := rules.ValidBaseType{ResolveObject: self.getObject}
		// rule.Validate(obj.Source())

		// typeAnalyzer := newTypeAnalyzer(self, obj.Source())
		// def := self.analyzeTypeStmt(obj.Source())
		// if def != nil {
		// 	nexFile.Types = append(nexFile.Types, *def)
		// }
	}

	var err error
	var hashcode uint64
	hashcode, err = hashstructure.Hash(&nexFile, hashstructure.FormatV2, nil)
	nexFile.Id = fmt.Sprint(hashcode)
	if err != nil {
		panic(err)
	}
	self.files = append(self.files, nexFile)
}

// getObject under the hood calls FindOjbect on self.currLocalScope and reports any error if any
func (self *Analyzer) getObject(decl *parser.DeclStmt) *scope.Object {
	name, alias := decl.Format()
	obj, needAlias := self.currLocalScope.FindObject(name, alias)
	if obj == nil {
		if needAlias {
			self.errors.Push(analyzer_error.ErrNeedAlias{}, decl.Pos)
		} else {
			self.errors.Push(analyzer_error.ErrTypeNotFound{name, alias}, decl.Pos)
		}
	} else {
		// todo: may this check if obj is Base and don't allow it to use
		return obj
	}

	return nil
}
