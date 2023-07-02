package analyzer

import (
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// Analyzer takes a linked list of built scopes and analyzes them syntactically.
// Also, if analysis succeed, a definition is built
type Analyzer struct {
	scopes []*scope.Scope
	errors AnalyzerErrorCollection

	currScope      *scope.Scope
	currLocalScope *scope.LocalScope
	currTypeId     string

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

	defaultValueValidField := &DefaultValueValidField{}
	nonNullableUnion := &NonNullableUnionField{}
	subsequentFieldIndex := &SubsequentFieldIndex{}
	uniqueDefaultValue := &UniqueDefaultValue{}
	uniqueFieldIndex := &UniqueFieldIndex{}
	uniqueFieldName := &UniqueFieldName{}
	validBaseType := &ValidBaseType{}
	validFieldType := &ValidFieldType{}
	validListArguments := &ValidListArguments{}
	validMapArguments := &ValidMapArguments{}
	validMapKey := &ValidMapKey{}

	defaultRules[defaultValueValidField.Key()] = defaultValueValidField
	defaultRules[nonNullableUnion.Key()] = nonNullableUnion
	defaultRules[subsequentFieldIndex.Key()] = subsequentFieldIndex
	defaultRules[uniqueDefaultValue.Key()] = uniqueDefaultValue
	defaultRules[uniqueFieldIndex.Key()] = uniqueFieldIndex
	defaultRules[uniqueFieldName.Key()] = uniqueFieldName
	defaultRules[validBaseType.Key()] = validBaseType
	defaultRules[validFieldType.Key()] = validFieldType
	defaultRules[validListArguments.Key()] = validListArguments
	defaultRules[validMapArguments.Key()] = validMapArguments
	defaultRules[validMapKey.Key()] = validMapKey
}

func NewAnalyzer(scopes []*scope.Scope) *Analyzer {
	analyzer := &Analyzer{
		scopes: scopes,
		errors: make([]*AnalyzerError, 0),
		rules:  defaultRules,
	}

	return analyzer
}

// Analyze starts analyzing and records any error encountered
func (self *Analyzer) Analyze() {
	for _, scope := range self.scopes {
		self.analyzeScope(scope)
	}
}

func (self *Analyzer) HasAnalysisErrors() bool {
	return len(self.errors) > 0
}

func (self *Analyzer) Errors() *AnalyzerErrorCollection {
	return &self.errors
}

func (self *Analyzer) analyzeScope(s *scope.Scope) {
	self.currScope = s
	for _, localScope := range *s.LocalScopes() {
		self.analyzeLocalScope(localScope)
	}
}

func (self *Analyzer) analyzeLocalScope(ls *scope.LocalScope) {
	self.currLocalScope = ls

	for _, obj := range *ls.Objects() {
		self.currTypeId = obj.Id

		for _, rule := range self.rules {
			context := &AnalyzerContext{
				scope:  ls,
				errors: NewAnalyzerErrorCollection(),
			}
			rule.Analyze(context)
			self.errors = append(self.errors, *context.Errors()...)
		}
	}
}
