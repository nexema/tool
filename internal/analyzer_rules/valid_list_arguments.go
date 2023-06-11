package analyzer_rules

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/analyzer"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// ValidListArguments checks if the value type defined as list contains exactly one type argument and also it is a valid Nexema value type
type ValidListArguments struct{}

func (self ValidListArguments) Analyze(context *analyzer.AnalyzerContext) {
	context.RunOver(func(object *scope.Object, source *parser.TypeStmt) {
		for _, stmt := range source.Fields {

			if !stmt.ValueType.Is(definition.List) {
				continue
			}

			if len(stmt.ValueType.Args) != 1 {
				context.ReportError(errInvalidListArgumentsLen{Given: len(stmt.ValueType.Args)}, stmt.ValueType.Pos)
				continue
			}

			typeArgument := stmt.ValueType.Args[0]
			verifyFieldType(&typeArgument, context, object) // maybe there is a cleaner way?
		}
	})
}

func (self ValidListArguments) Throws() analyzer.RuleThrow {
	return analyzer.Error
}

func (self ValidListArguments) Key() string {
	return "valid-list-arguments"
}

type errInvalidListArgumentsLen struct {
	Given int
}

func (e errInvalidListArgumentsLen) Message() string {
	return fmt.Sprintf("list type expects exactly one type argument, given %d", e.Given)
}
