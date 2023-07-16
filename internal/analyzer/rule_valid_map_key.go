package analyzer

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// ValidMapKey checks if the key type of the value type defined as map is a non nullable string, bool or (u)(var)int(8|16|32|64)
type ValidMapKey struct{}

var availableKeyTypes = map[definition.ValuePrimitive]bool{
	definition.String:  true,
	definition.Boolean: true,
	definition.Uint:    true,
	definition.Uint8:   true,
	definition.Uint16:  true,
	definition.Uint32:  true,
	definition.Uint64:  true,
	definition.Int:     true,
	definition.Int8:    true,
	definition.Int16:   true,
	definition.Int32:   true,
	definition.Int64:   true,
}

func (self ValidMapKey) Analyze(context *AnalyzerContext) {
	context.RunOver(func(object *scope.Object, source *parser.TypeStmt) {
		for _, stmt := range source.Fields {

			if stmt.ValueType == nil || !stmt.ValueType.Is(definition.Map) {
				continue
			}

			if len(stmt.ValueType.Args) != 2 {
				continue // checked by other rule
			}

			keyArgument := stmt.ValueType.Args[0]
			primitive, ok := definition.ParsePrimitive(keyArgument.Token.Literal)
			if !ok {
				context.ReportError(errInvalidMapKey{}, keyArgument.Pos)
			} else if _, ok := availableKeyTypes[primitive]; !ok || ok && keyArgument.Nullable {
				context.ReportError(errInvalidMapKey{primitive}, keyArgument.Pos)
			}
		}
	})
}

func (self ValidMapKey) Throws() RuleThrow {
	return Error
}

func (self ValidMapKey) Key() string {
	return "valid-map-key"
}

type errInvalidMapKey struct {
	GivenType definition.ValuePrimitive
}

func (e errInvalidMapKey) Message() string {
	msg := "map expects its key to be a non nullable string, bool or any type of int"
	if len(e.GivenType) > 0 {
		msg = fmt.Sprintf("%s, given %s", msg, e.GivenType)
	}

	return msg
}
