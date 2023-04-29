package rules

import (
	"fmt"

	analyzer_error "tomasweigenast.com/nexema/tool/internal/analyzer/error"
	"tomasweigenast.com/nexema/tool/internal/parser"
)

// DuplicatedFieldName checks if field names defined in a struct are not duplicated
type DuplicatedFieldName struct {
	group *AnalyzerRuleCollection
}

func (self DuplicatedFieldName) Validate(value any) []analyzer_error.AnalyzerErrorKind {
	typeStmt := value.(*parser.TypeStmt)

	errs := make([]analyzer_error.AnalyzerErrorKind, 0)
	check := map[string]bool{}
	for _, stmt := range typeStmt.Fields {
		fieldName := stmt.Name.Token.Literal
		if _, ok := check[fieldName]; !ok {
			errs = append(errs, errDuplicatedFieldName{FieldName: fieldName})
		} else {
			check[fieldName] = true
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

type errDuplicatedFieldName struct {
	FieldName string
}

func (e errDuplicatedFieldName) Message() string {
	return fmt.Sprintf("field %s already defined", e.FieldName)
}
