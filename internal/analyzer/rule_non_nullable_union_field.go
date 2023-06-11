package analyzer

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
	"tomasweigenast.com/nexema/tool/internal/token"
)

// NonNullableUnionField checks if the fields defined in an union are not nullable
type NonNullableUnionField struct{}

func (self NonNullableUnionField) Analyze(context *AnalyzerContext) {
	context.RunOver(func(object *scope.Object, source *parser.TypeStmt) {
		if source.Modifier != token.Union {
			return
		}

		for _, stmt := range source.Fields {
			if stmt.ValueType == nil {
				continue
			}

			if stmt.ValueType.Nullable {
				context.ReportError(errNonNullableUnionField{}, stmt.Name.Pos)
			}
		}
	})
}

func (self NonNullableUnionField) Throws() RuleThrow {
	return Error
}

func (self NonNullableUnionField) Key() string {
	return "non-nullable-union-field"
}

type errNonNullableUnionField struct {
}

func (e errNonNullableUnionField) Message() string {
	return fmt.Sprintf("fields must not be nullable when declared in an union")
}
