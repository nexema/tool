package rules

import (
	"tomasweigenast.com/nexema/tool/internal/analyzer"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// ValidFieldType checks if the value type of a field is a valid Nexema type or an imported type
type ValidFieldType struct{}

func (self ValidFieldType) Analyze(context *analyzer.AnalyzerContext) {
	context.RunOver(func(object *scope.Object, source *parser.TypeStmt) {
		for _, stmt := range source.Fields {

			// this should not happen
			if stmt.ValueType == nil {
				panic("this should not happen, field does not have a defined value type?")
			}

			typeName, _ := stmt.ValueType.Format()
			_, valid := definition.ParsePrimitive(typeName)

			if valid {
				continue
			} else {
				context.GetObject(stmt.ValueType)
			}
		}
	})
}

func (self ValidFieldType) Throws() analyzer.RuleThrow {
	return analyzer.Error
}

func (self ValidFieldType) Key() string {
	return "valid-value-type"
}
