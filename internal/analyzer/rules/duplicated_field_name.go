package rules

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/analyzer"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// DuplicatedFieldName checks if field names defined in a struct are not duplicated
type DuplicatedFieldName struct{}

func (self DuplicatedFieldName) Analyze(context *analyzer.AnalyzerContext) {
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

type errDuplicatedFieldName struct {
	FieldName string
}

func (e errDuplicatedFieldName) Message() string {
	return fmt.Sprintf("field %s already defined", e.FieldName)
}
