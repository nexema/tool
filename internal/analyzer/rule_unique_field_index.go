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
		check := map[int]bool{}
		for _, stmt := range source.Fields {

			// skip fields that does not have an index
			if stmt.Index == nil {
				continue
			}

			fieldIndex, _ := strconv.Atoi(stmt.Index.Token.Literal)
			if _, ok := check[fieldIndex]; ok {
				context.ReportError(errDuplicatedFieldIndex{FieldIndex: fieldIndex}, stmt.Index.Pos)
			} else {
				check[fieldIndex] = true
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
}

func (e errDuplicatedFieldIndex) Message() string {
	return fmt.Sprintf("field index %d already in use", e.FieldIndex)
}
