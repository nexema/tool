package analyzer

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
	"tomasweigenast.com/nexema/tool/internal/utils"
)

// SubsequentFieldIndex checks if field indexes are subsequent and start from 0
type SubsequentFieldIndex struct{}

func (self SubsequentFieldIndex) Analyze(context *AnalyzerContext) {
	context.RunOver(func(object *scope.Object, source *parser.TypeStmt) {
		if len(source.Fields) == 0 {
			return
		}

		lastIdx := 0
		for i, stmt := range source.Fields {

			if i == 0 {
				if stmt.Index != nil && stmt.Index.Token.Literal != "0" {
					context.ReportError(errFirstFieldNotZero{Given: utils.ToInt(stmt.Index.Token.Literal)}, source.Fields[0].Index.Pos)
					return
				} else {
					continue
				}
			}

			if stmt.Index == nil {
				lastIdx++
				continue
			}

			fieldIndex := utils.ToInt(stmt.Index.Token.Literal)
			if lastIdx+1 != fieldIndex {
				context.ReportError(errNonSubsequentFieldIndex{FieldIndex: fieldIndex}, stmt.Index.Pos)
			} else {
				lastIdx = fieldIndex
			}
		}
	})
}

func (self SubsequentFieldIndex) Throws() RuleThrow {
	return Error
}

func (self SubsequentFieldIndex) Key() string {
	return "unique-field-index"
}

type errNonSubsequentFieldIndex struct {
	FieldIndex int
}

type errFirstFieldNotZero struct {
	Given int
}

func (e errNonSubsequentFieldIndex) Message() string {
	return fmt.Sprintf("field index %d already in use", e.FieldIndex)
}

func (e errFirstFieldNotZero) Message() string {
	return fmt.Sprintf("when explicitly declaring indexes for fields, the first one must be zero, given %d", e.Given)
}
