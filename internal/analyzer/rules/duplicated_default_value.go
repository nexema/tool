package rules

import (
	"fmt"

	analyzer_error "tomasweigenast.com/nexema/tool/internal/analyzer/error"
	"tomasweigenast.com/nexema/tool/internal/parser"
)

// DuplicatedDefaultValue checks if the are no duplicated default values in a struct
type DuplicatedDefaultValue struct {
	group *AnalyzerRuleCollection
}

func (self DuplicatedDefaultValue) Validate(value any) []analyzer_error.AnalyzerErrorKind {
	typeStmt := value.(*parser.TypeStmt)

	if typeStmt.Defaults == nil || len(typeStmt.Defaults) == 0 {
		return nil
	}

	check := map[string]any{}
	errs := make([]analyzer_error.AnalyzerErrorKind, 0)
	for _, stmt := range typeStmt.Defaults {
		key := stmt.Left.Token.Literal
		if _, ok := check[key]; !ok {
			errs = append(errs, errDuplicatedDefaultValue{FieldName: key})
		} else {
			check[key] = true
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

type errDuplicatedDefaultValue struct {
	FieldName string
}

func (e errDuplicatedDefaultValue) Message() string {
	return fmt.Sprintf("default value for field %s already defined", e.FieldName)
}
