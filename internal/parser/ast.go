package parser

import (
	"fmt"
	"strings"

	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
)

type Ast struct {
	File           *reference.File
	UseStatements  []UseStmt
	TypeStatements []TypeStmt
}

type CommentStmt struct {
	Token token.Token
	Pos   reference.Pos
}

type UseStmt struct {
	Token token.Token // The "use" token
	Path  LiteralStmt
	Alias *IdentStmt
}

type AnnotationStmt struct {
	Token     token.Token
	Assigment AssignStmt
	Pos       reference.Pos
}

type IdentStmt struct {
	Token token.Token
	Pos   reference.Pos
}

type DeclStmt struct {
	Token    token.Token
	Pos      reference.Pos
	Args     []DeclStmt
	Alias    *IdentStmt
	Nullable bool
}

type AssignStmt struct {
	Token token.Token // the "=" token
	Left  IdentStmt
	Right LiteralStmt
	Pos   reference.Pos
}

type LiteralStmt struct {
	Token token.Token
	Kind  LiteralKind
	Pos   reference.Pos
}

type LiteralKind interface {
	Literal() string
	Value() interface{}
}

type BooleanLiteral struct {
	value bool
}

type IntLiteral struct {
	value int64
}

type FloatLiteral struct {
	value float64
}

type StringLiteral struct {
	value string
}

type ListLiteral []LiteralStmt

type MapLiteral []MapEntry

type MapEntry struct {
	Key, Value LiteralStmt
}

type TypeStmt struct {
	Name          IdentStmt
	Modifier      token.TokenKind
	BaseType      *DeclStmt
	Documentation []CommentStmt
	Annotations   []AnnotationStmt
	Fields        []FieldStmt
	Defaults      []AssignStmt
}

type FieldStmt struct {
	Index         *IdentStmt
	Name          IdentStmt
	ValueType     *DeclStmt
	Documentation []CommentStmt
	Annotations   []AnnotationStmt
}

func (self BooleanLiteral) Literal() string {
	return fmt.Sprint(self.value)
}

func (self BooleanLiteral) Value() interface{} {
	return self.value
}

func (self IntLiteral) Literal() string {
	return fmt.Sprint(self.value)
}

func (self IntLiteral) Value() interface{} {
	return self.value
}

func (self FloatLiteral) Literal() string {
	return fmt.Sprint(self.value)
}

func (self FloatLiteral) Value() interface{} {
	return self.value
}

func (self StringLiteral) Literal() string {
	return fmt.Sprintf(`"%s"`, self.value)
}

func (self StringLiteral) Value() interface{} {
	return self.value
}

func (self ListLiteral) Literal() string {
	return fmt.Sprintf("[%v]", strings.Join(mapArray(self, func(elem LiteralStmt) string {
		return fmt.Sprint(elem.Kind.Literal())
	}), ", "))
}

func (self ListLiteral) Value() interface{} {
	return mapArray(self, func(elem LiteralStmt) interface{} {
		return elem.Kind.Value()
	})
}

func (self MapLiteral) Literal() string {
	out := make([]string, len(self))
	for _, v := range self {
		out = append(out, fmt.Sprintf("(%v: %v)", v.Key.Kind.Literal(), v.Value.Kind.Literal()))
	}

	return fmt.Sprintf("[%s]", strings.Join(out, ", "))
}

func (self MapLiteral) Value() interface{} {
	out := make(map[interface{}]interface{}, len(self))
	for _, v := range self {
		out[v.Key.Kind.Value()] = v.Value.Kind.Value()
	}

	return out
}

func (self *DeclStmt) Format() (name, alias string) {
	name = self.Token.Literal
	if self.Alias != nil {
		alias = self.Alias.Token.Literal
	}

	return
}

func (self *DeclStmt) Is(primitive definition.ValuePrimitive) bool {
	name := self.Token.Literal
	prim, ok := definition.ParsePrimitive(name)
	if ok && primitive == prim {
		return true
	}

	return false
}

func mapArray[T any, O any](in []T, f func(T) O) []O {
	out := make([]O, len(in))
	for i, elem := range in {
		out[i] = f(elem)
	}

	return out
}

func MakeBooleanLiteral(v bool) BooleanLiteral {
	return BooleanLiteral{v}
}

func MakeStringLiteral(v string) StringLiteral {
	return StringLiteral{v}
}

func MakeIntLiteral(v int64) IntLiteral {
	return IntLiteral{v}
}

func MakeFloatLiteral(v float64) FloatLiteral {
	return FloatLiteral{v}
}

func MakeListLiteral(values ...LiteralStmt) ListLiteral {
	return values
}

func MakeMapLiteral(values ...MapEntry) MapLiteral {
	return values
}

// UnwrapAlias returns the alias' literal or an empty string if alias is nil
func (self *UseStmt) UnwrapAlias() string {
	if self.Alias == nil {
		return ""
	}

	return self.Alias.Token.Literal
}
