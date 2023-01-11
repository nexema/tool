package internal

import "fmt"

// Ast represents the abstract syntax tree of a single Nexema file
type Ast struct {
	File    *File
	Imports *[]*ImportStmt
	Types   *[]*TypeStmt
}

// File represents the origin file which was used to build an Ast
type File struct {
	Name string // file name
	Pkg  string // package path, relative to nexema.yaml
}

// Comment represents a comment read on a file
type CommentStmt struct {
	Text      string // the comment's literal
	posStart  int
	posEnd    int
	lineStart int
	lineEnd   int
}

type ImportStmt struct {
	Path  *IdentifierStmt
	Alias *IdentifierStmt
}

type TypeStmt struct {
	Name          *IdentifierStmt
	Modifier      Token // Token_Struct, Token_Enum, Token_Union
	Metadata      *MapValueStmt
	Documentation *[]*CommentStmt
	Fields        *[]*FieldStmt
}

type FieldStmt struct {
	Index         ValueStmt
	Name          *IdentifierStmt
	ValueType     *ValueTypeStmt
	Metadata      *MapValueStmt
	DefaultValue  ValueStmt
	Documentation *[]*CommentStmt
}

type ValueTypeStmt struct {
	Ident         *IdentifierStmt
	Nullable      bool
	TypeArguments *[]*ValueTypeStmt
}

type IdentifierStmt struct {
	Lit   string
	Alias string // my_alias.EnumType
}

type ValueStmt interface {
	Kind() Primitive
	Value() interface{}
}

type PrimitiveValueStmt struct {
	RawValue  interface{}
	Primitive Primitive // primitives without list, map or type
}

// TypeValueStmt represents a value of an enum
type TypeValueStmt struct {
	TypeName *IdentifierStmt
	RawValue *IdentifierStmt
}

type MapValueStmt []*MapEntryStmt
type MapEntryStmt struct {
	Key   ValueStmt
	Value ValueStmt
}

func (m *MapValueStmt) add(stmt *MapEntryStmt) {
	(*m) = append((*m), stmt)
}

type ListValueStmt []ValueStmt

func (l *ListValueStmt) add(stmt ValueStmt) {
	(*l) = append(*l, stmt)
}

func (p *PrimitiveValueStmt) Kind() Primitive {
	return p.Primitive
}

func (*TypeValueStmt) Kind() Primitive {
	return Primitive_Type
}

func (*ListValueStmt) Kind() Primitive {
	return Primitive_List
}

func (*MapValueStmt) Kind() Primitive {
	return Primitive_Map
}

func (p *TypeValueStmt) Value() interface{} {
	if p.TypeName.Alias == "" {
		return fmt.Sprintf("%s.%s", p.TypeName.Lit, p.RawValue.Lit)
	}

	return fmt.Sprintf("%s.%s.%s", p.TypeName.Alias, p.TypeName.Lit, p.RawValue.Lit)
}

func (p *PrimitiveValueStmt) Value() interface{} {
	return p.RawValue
}

func (l *ListValueStmt) Value() interface{} {
	arr := make([]any, len(*l))
	for i, val := range *l {
		arr[i] = val.Value()
	}
	return arr
}

func (m *MapValueStmt) Value() interface{} {
	ma := make(map[any]any, len(*m))
	for _, val := range *m {
		key := val.Key.Value()
		value := val.Value.Value()
		ma[key] = value
	}
	return ma
}
