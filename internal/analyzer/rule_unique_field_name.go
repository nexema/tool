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
		for _, stmt := range source.Fields {
			fieldName := stmt.Name.Token.Literal
			if _, ok := check[fieldName]; ok {
				context.ReportError(errDuplicatedFieldName{FieldName: fieldName}, stmt.Name.Pos)
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
	FieldName string
}

func (e errDuplicatedFieldName) Message() string {
	return fmt.Sprintf("field %s already defined", e.FieldName)
}
