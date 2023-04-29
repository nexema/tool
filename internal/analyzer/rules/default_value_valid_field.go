package rules

import (
	"fmt"

	analyzer_error "tomasweigenast.com/nexema/tool/internal/analyzer/error"
	"tomasweigenast.com/nexema/tool/internal/parser"
)

// DefaultValueValidField checks if the field defined in a default value declaration exists
type DefaultValueValidField struct {
	group *AnalyzerRuleCollection
}

func (self DefaultValueValidField) Validate(value any) []analyzer_error.AnalyzerErrorKind {
	typeStmt := value.(*parser.TypeStmt)

	if typeStmt.Defaults == nil || len(typeStmt.Defaults) == 0 {
		return nil
	}

	errs := make([]analyzer_error.AnalyzerErrorKind, 0)
	for _, stmt := range typeStmt.Defaults {
		fieldName := stmt.Left.Token.Literal

		var fieldStmt *parser.FieldStmt
		for _, field := range typeStmt.Fields {
			if field.Name.Token.Literal == fieldName {
				fieldStmt = &field
			}
		}

		if fieldStmt == nil {
			errs = append(errs, errDefaultValueValidField{FieldName: fieldName})
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

type errDefaultValueValidField struct {
	FieldName string
}

func (e errDefaultValueValidField) Message() string {
	return fmt.Sprintf("field %s not defined", e.FieldName)
}
