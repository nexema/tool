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

type BooleanKind struct {
	value bool
}

type IntKind struct {
	value int64
}

type FloatKind struct {
	value float64
}

type StringKind struct {
	value string
}

type ListKind []LiteralStmt

type MapKind map[LiteralStmt]LiteralStmt

type TypeStmt struct {
	Name          IdentStmt
	Modifier      token.Token
	BaseType      *DeclStmt
	Documentation []CommentStmt
	Annotations   []AnnotationStmt
	Fields        []FieldStmt
	Defaults      []AssignStmt
}

type FieldStmt struct {
	Index         *IdentStmt
	Name          IdentStmt
	ValueType     DeclStmt
	Documentation []CommentStmt
	Annotations   []AnnotationStmt
}

func (self BooleanKind) Literal() string {
	return fmt.Sprint(self.value)
}

func (self BooleanKind) Value() interface{} {
	return self.Value
}

func (self IntKind) Literal() string {
	return fmt.Sprint(self.value)
}

func (self IntKind) Value() interface{} {
	return self.Value
}

func (self FloatKind) Literal() string {
	return fmt.Sprint(self.value)
}

func (self FloatKind) Value() interface{} {
	return self.Value
}

func (self StringKind) Literal() string {
	return self.value
}

func (self StringKind) Value() interface{} {
	return self.Value
}

func (self ListKind) Literal() string {
	return fmt.Sprintf("[%v]", strings.Join(utils.MapArray(self, func(elem LiteralStmt) string {
		return fmt.Sprint(elem.Kind.Value())
	}), ", "))
}

func (self ListKind) Value() interface{} {
	return utils.MapArray(self, func(elem LiteralStmt) interface{} {
		return elem.Kind.Value()
	})
}

func (self MapKind) Literal() string {
	out := make([]string, len(self))
	for k, v := range self {
		out = append(out, fmt.Sprintf("(%v: %v)", k.Kind.Value(), v.Kind.Value()))
	}

	return fmt.Sprintf("[%s]", strings.Join(out, ", "))
}
