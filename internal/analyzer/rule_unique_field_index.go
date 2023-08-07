package analyzer

import (
	"fmt"
	"strconv"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// UniqueFieldIndex checks if field indexes defined in a struct are not duplicated
type UniqueFieldIndex struct{}

func (self UniqueFieldIndex) Analyze(context *AnalyzerContext) {
	context.RunOver(func(object *scope.Object, source *parser.TypeStmt) {
		// true value means a field index from a base type, false from the current's type
		check := map[int]bool{}

		if source.BaseType != nil {
			baseTypeObj := context.GetObject(source.BaseType)
			baseTypeIndex := 0
			for _, field := range baseTypeObj.Source().Fields {
				if field.Index != nil {
					baseTypeIndex, _ = strconv.Atoi(field.Index.Token.Literal)
				}

				check[baseTypeIndex] = true
				baseTypeIndex++
			}
		}

		for _, stmt := range source.Fields {

			// skip fields that does not have an index
			if stmt.Index == nil {
				continue
			}

			fieldIndex, _ := strconv.Atoi(stmt.Index.Token.Literal)
			if value, ok := check[fieldIndex]; ok {
				context.ReportError(errDuplicatedFieldIndex{FieldIndex: fieldIndex, IsFromBase: value}, stmt.Index.Pos)
			} else {
				check[fieldIndex] = false
			}
		}
	})
}

func (self UniqueFieldIndex) Throws() RuleThrow {
	return Error
}

func (self UniqueFieldIndex) Key() string {
	return "unique-field-index"
}

type errDuplicatedFieldIndex struct {
	FieldIndex int
	IsFromBase bool // If true, indicates the field conflict comes from a field defined in a base type
}

func (e errDuplicatedFieldIndex) Message() string {
	if e.IsFromBase {
		return fmt.Sprintf("field index %d already used by a field in the base type", e.FieldIndex)
	} else {
		return fmt.Sprintf("field index %d already used by a field", e.FieldIndex)
	}
}
