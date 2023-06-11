package analyzer_rules

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
			verifyFieldType(stmt.ValueType, context, object)
		}
	})
}

func verifyFieldType(stmt *parser.DeclStmt, context *analyzer.AnalyzerContext, object *scope.Object) {

	typeName, _ := stmt.Format()
	_, valid := definition.ParsePrimitive(typeName)

	if valid {
		return
	} else {
		obj := context.GetObject(stmt)
		if object != nil && obj != nil && obj.Id == object.Id {
			context.ReportError(errTypeNotAllowed{}, stmt.Pos)
		}
	}
}

func (self ValidFieldType) Throws() analyzer.RuleThrow {
	return analyzer.Error
}

func (self ValidFieldType) Key() string {
	return "valid-value-type"
}

type errTypeNotAllowed struct{}

func (errTypeNotAllowed) Message() string {
	return "the value type of a field cannot be the same as the type where the field is being declared"
}
