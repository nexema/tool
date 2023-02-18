package parser

import (
	"fmt"
	"strings"

	"tomasweigenast.com/nexema/tool/token"
	"tomasweigenast.com/nexema/tool/tokenizer"
	"tomasweigenast.com/nexema/tool/utils"
)

type File struct {
	Path     string
	FileName string
}

type Ast struct {
	File           *File
	UseStatements  []UseStmt
	TypeStatements []TypeStmt
}

type CommentStmt struct {
	Token token.Token
	Pos   tokenizer.Pos
}

type UseStmt struct {
	Token token.Token // The "use" token
	Path  LiteralStmt
	Alias *IdentStmt
}

type AnnotationStmt struct {
	Token     token.Token
	Assigment AssignStmt
	Pos       tokenizer.Pos
}

type IdentStmt struct {
	Token token.Token
	Pos   tokenizer.Pos
}

type DeclStmt struct {
	Token    token.Token
	Pos      tokenizer.Pos
	Args     []DeclStmt
	Alias    *IdentStmt
	Nullable bool
}

type AssignStmt struct {
	Token token.Token // the "=" token
	Left  IdentStmt
	Right LiteralStmt
	Pos   tokenizer.Pos
}

type LiteralStmt struct {
	Token token.Token
	Kind  LiteralKind
	Pos   tokenizer.Pos
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
	return self.Value
}

func (self IntLiteral) Literal() string {
	return fmt.Sprint(self.value)
}

func (self IntLiteral) Value() interface{} {
	return self.Value
}

func (self FloatLiteral) Literal() string {
	return fmt.Sprint(self.value)
}

func (self FloatLiteral) Value() interface{} {
	return self.Value
}

func (self StringLiteral) Literal() string {
	return self.value
}

func (self StringLiteral) Value() interface{} {
	return self.Value
}

func (self ListLiteral) Literal() string {
	return fmt.Sprintf("[%v]", strings.Join(utils.MapArray(self, func(elem LiteralStmt) string {
		return fmt.Sprint(elem.Kind.Value())
	}), ", "))
}

func (self ListLiteral) Value() interface{} {
	return utils.MapArray(self, func(elem LiteralStmt) interface{} {
		return elem.Kind.Value()
	})
}

func (self MapLiteral) Literal() string {
	out := make([]string, len(self))
	for _, v := range self {
		out = append(out, fmt.Sprintf("(%v: %v)", v.Key.Kind.Value(), v.Value.Kind.Value()))
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
