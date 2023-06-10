package rules

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/analyzer"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
	"tomasweigenast.com/nexema/tool/internal/token"
)

// ValidBaseType checks if the extended type in a struct exists and is a valid Base type.
type ValidBaseType struct{}

func (self ValidBaseType) Analyze(context *analyzer.AnalyzerContext) {
	context.RunOver(func(object *scope.Object, source *parser.TypeStmt) {
		src := object.Source()
		if src.BaseType == nil {
			return
		}

		baseType := context.GetObject(src.BaseType)
		if baseType != nil && baseType.Source().Modifier != token.Base {
			context.ReportError(errWrongBaseType{TypeName: baseType.Name}, src.BaseType.Pos)
		}
	})
}

func (self ValidBaseType) Throws() analyzer.RuleThrow {
	return analyzer.Error
}

type errWrongBaseType struct {
	TypeName string
}

func (e errWrongBaseType) Message() string {
	return fmt.Sprintf("%s is not a valid Base type, it cannot be extended", e.TypeName)
}
