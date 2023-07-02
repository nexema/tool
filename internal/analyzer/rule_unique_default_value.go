package analyzer

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// UniqueDefaultValue checks if the are no duplicated default values in a struct
type UniqueDefaultValue struct{}

func (self UniqueDefaultValue) Analyze(context *AnalyzerContext) {
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

func (self UniqueDefaultValue) Throws() RuleThrow {
	return Error
}

func (self UniqueDefaultValue) Key() string {
	return "unique-default-value"
}

type errDuplicatedDefaultValue struct {
	FieldName string
}

func (e errDuplicatedDefaultValue) Message() string {
	return fmt.Sprintf("default value for field %s already defined", e.FieldName)
}