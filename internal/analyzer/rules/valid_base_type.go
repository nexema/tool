package rules

import (
	"fmt"

	analyzer_error "tomasweigenast.com/nexema/tool/internal/analyzer/error"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/token"
)

// ValidBaseType checks if the base type in a struct exists and is a valid Base type.
type ValidBaseType struct {
	group *AnalyzerRuleCollection
}

func (self ValidBaseType) Validate(value any) []analyzer_error.AnalyzerErrorKind {
	typeStmt := value.(*parser.TypeStmt)

	if typeStmt.BaseType == nil {
		return nil
	}

	obj := self.group.ResolveObject(typeStmt.BaseType)
	if obj != nil && obj.Source().Modifier != token.Base {
		return []analyzer_error.AnalyzerErrorKind{errWrongBaseType{TypeName: obj.Name}}
	}

	return nil
}

type errWrongBaseType struct {
	TypeName string
}

func (e errWrongBaseType) Message() string {
	return fmt.Sprintf("%s is not a valid Base type, it cannot be extended", e.TypeName)
}
