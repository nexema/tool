package analyzer

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// DefaultValueValidField checks if the field defined in a default value declaration exists
type DefaultValueValidField struct{}

func (self DefaultValueValidField) Analyze(context *AnalyzerContext) {
	context.RunOver(func(object *scope.Object, source *parser.TypeStmt) {
		if source.Defaults == nil || len(source.Defaults) == 0 {
			return
		}

		for _, stmt := range source.Defaults {
			fieldName := stmt.Left.Token.Literal

			var fieldStmt *parser.FieldStmt
			for _, field := range source.Fields {
				if field.Name.Token.Literal == fieldName {
					fieldStmt = &field
				}
			}

			if fieldStmt == nil {
				context.ReportError(errDefaultValueValidField{FieldName: fieldName}, stmt.Pos)
			}
		}
	})
}

func (self DefaultValueValidField) Throws() RuleThrow {
	return Error
}

func (self DefaultValueValidField) Key() string {
	return "default-value-valid-field"
}

type errDefaultValueValidField struct {
	FieldName string
}

func (e errDefaultValueValidField) Message() string {
	return fmt.Sprintf("field %s not defined", e.FieldName)
}
