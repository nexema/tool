package analyzer

import (
	"fmt"
	"path"

	"github.com/mitchellh/hashstructure/v2"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// Analyzer takes a linked list of built scopes and analyzes them syntactically.
// Also, if analysis succeed, a definition is built
type Analyzer struct {
	scopes []*scope.Scope
	errors *AnalyzerErrorCollection

	currScope      *scope.Scope
	currLocalScope *scope.LocalScope
	currTypeId     string
	files          []definition.NexemaFile

	rules map[string]AnalyzerRule // the list of rules. key is the name of the rule and the value the actual rule executor
}

// RuleThrow identifies the type of error thrown by a rule
type RuleThrow int8

const (
	Error   RuleThrow = 1 // Error prevents the project to be compiled
	Warning RuleThrow = 2 // Warning is shown as an improvement
)

// AnalyzerRule is the base interface for every rule.
type AnalyzerRule interface {
	Analyze(context *AnalyzerContext)

	// Throws indicates what kind of error is thrown by the rule
	Throws() RuleThrow

	// Key returns an unique identifier for the rule
	Key() string
}

var defaultRules map[string]AnalyzerRule

func init() {
	defaultRules = make(map[string]AnalyzerRule)

	// uniqueFieldName := analyzer_rules.UniqueFieldName{}

}

func NewAnalyzer(scopes []*scope.Scope) *Analyzer {
	analyzer := &Analyzer{
		scopes: scopes,
		errors: NewAnalyzerErrorCollection(),
		files:  make([]definition.NexemaFile, 0),
		rules:  defaultRules,
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

		for _, rule := range self.rules {
			context := &AnalyzerContext{
				scope:  ls,
				errors: NewAnalyzerErrorCollection(),
			}
			rule.Analyze(context)
		}
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
