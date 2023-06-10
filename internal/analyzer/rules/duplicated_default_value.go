package rules

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/analyzer"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// DuplicatedDefaultValue checks if the are no duplicated default values in a struct
type DuplicatedDefaultValue struct{}

func (self DuplicatedDefaultValue) Analyze(context *analyzer.AnalyzerContext) {
	context.RunOver(func(object *scope.Object, source *parser.TypeStmt) {
		if source.Defaults == nil || len(source.Defaults) == 0 {
			return
		}

		check := map[string]any{}
		for _, stmt := range source.Defaults {
			key := stmt.Left.Token.Literal
			if _, ok := check[key]; ok {
				context.ReportError(errDuplicatedDefaultValue{FieldName: key}, stmt.Pos)
			} else {
				check[key] = true
			}
		}
	})
}

func (self DuplicatedDefaultValue) Throws() analyzer.RuleThrow {
	return analyzer.Error
}

type errDuplicatedDefaultValue struct {
	FieldName string
}

func (e errDuplicatedDefaultValue) Message() string {
	return fmt.Sprintf("default value for field %s already defined", e.FieldName)
}
