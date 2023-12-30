package analysis

import (
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/token"
)

// typeContext contains information about the current type that is being analyzed
type typeContext struct {
	statement             *parser.TypeStatement // the type
	fieldNames            map[string]bool       // a map to keep track of field names for duplicates
	fieldIndexes          map[int64]bool        // a map to keep track of used field indexes
	fieldsRead            bool                  // A flag that indicates if it end reading field statements
	defaultsRead          bool                  // a flag that indicates if it read a defaults statement
	nextAvailableFieldIdx int64                 // the last index read or automatically deduced for a field
}

func (tc *typeContext) Statement() parser.Statement {
	return tc.statement
}

func (tc *typeContext) IsUnion() bool {
	if tc.statement.Modifier == nil {
		return false
	}

	return tc.statement.Modifier.Token.Kind == token.Union
}

func (tc *typeContext) IsEnum() bool {
	if tc.statement.Modifier == nil {
		return false
	}
	return tc.statement.Modifier.Token.Kind == token.Enum
}
