package analyzer

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// UniqueFieldName checks if field names defined in a struct are not duplicated
type UniqueFieldName struct{}

func (self UniqueFieldName) Analyze(context *AnalyzerContext) {
	context.RunOver(func(object *scope.Object, source *parser.TypeStmt) {
		check := map[string]bool{}
		baseFields := map[string]bool{}
		if source.BaseType != nil {
			base := context.GetObject(source.BaseType)
			if base != nil {
				for _, baseField := range base.Source().Fields {
					baseFields[baseField.Name.Token.Literal] = true
				}
			}
		}

		for _, stmt := range source.Fields {
			fieldName := stmt.Name.Token.Literal
			if _, ok := check[fieldName]; ok {
				context.ReportError(errDuplicatedFieldName{FieldName: fieldName}, stmt.Name.Pos)
			} else if _, ok := baseFields[fieldName]; ok {
				context.ReportError(errDuplicatedFieldName{FieldName: fieldName, IsFromBase: true}, stmt.Name.Pos)
			} else {
				check[fieldName] = true
			}
		}
	})
}

func (self UniqueFieldName) Throws() RuleThrow {
	return Error
}

func (self UniqueFieldName) Key() string {
	return "unique-field-name"
}

type errDuplicatedFieldName struct {
	FieldName  string
	IsFromBase bool
}

func (e errDuplicatedFieldName) Message() string {
	if e.IsFromBase {
		return fmt.Sprintf("field %q is already defined in the base type", e.FieldName)
	} else {
		return fmt.Sprintf("field %q already defined", e.FieldName)
	}
}
