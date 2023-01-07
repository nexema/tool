package internal

// Ast represents the abstract syntax tree of a Nexema file
type Ast struct {
	imports *[]*ImportStmt
	types   *[]*TypeStmt
}

// Comment represents a comment read on a file
type CommentStmt struct {
	text      string // the comment's literal
	posStart  int
	posEnd    int
	lineStart int
	lineEnd   int
}

type ImportStmt struct {
	path  *IdentifierStmt
	alias *IdentifierStmt
}

type TypeStmt struct {
	name          *IdentifierStmt
	modifier      Token // Token_Struct, Token_Enum, Token_Union
	metadata      *MapValueStmt
	documentation *[]*CommentStmt
	fields        *[]*FieldStmt
}

type FieldStmt struct {
	index         ValueStmt
	name          *IdentifierStmt
	valueType     *ValueTypeStmt
	metadata      *MapValueStmt
	defaultValue  ValueStmt
	documentation *[]*CommentStmt
}

type ValueTypeStmt struct {
	ident         *IdentifierStmt
	nullable      bool
	typeArguments *[]*ValueTypeStmt
}

type IdentifierStmt struct {
	lit   string
	alias string // my_alias.EnumType
}

type ValueStmt interface {
	Kind() Primitive
}

type PrimitiveValueStmt struct {
	value interface{}
	kind  Primitive // primitives without list, map or type
}

// TypeValueStmt represents a value of an enum
type TypeValueStmt struct {
	typeName *IdentifierStmt
	value    *IdentifierStmt
}

type MapValueStmt []*MapEntryStmt
type MapEntryStmt struct {
	key   ValueStmt
	value ValueStmt
}

func (m *MapValueStmt) add(stmt *MapEntryStmt) {
	(*m) = append((*m), stmt)
}

type ListValueStmt []ValueStmt

func (l *ListValueStmt) add(stmt ValueStmt) {
	(*l) = append(*l, stmt)
}

func (p *PrimitiveValueStmt) Kind() Primitive {
	return p.kind
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
