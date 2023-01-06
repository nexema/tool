package internal

// Ast represents the abstract syntax tree of a Nexema file
type Ast struct {
	imports *[]*ImportStmt
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
	Name     *IdentifierStmt
	Modifier Token // Token_Struct, Token_Enum, Token_Union
	Metadata *MapStmt
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

type MapStmt []*MapEntryStmt
type MapEntryStmt struct {
	key   ValueStmt
	value ValueStmt
}

func (m *MapStmt) add(stmt *MapEntryStmt) {
	(*m) = append((*m), stmt)
}

type ListStmt []ValueStmt

func (l *ListStmt) add(stmt ValueStmt) {
	(*l) = append(*l, stmt)
}

func (p *PrimitiveValueStmt) Kind() Primitive {
	return p.kind
}

func (*TypeValueStmt) Kind() Primitive {
	return Primitive_Type
}

func (*ListStmt) Kind() Primitive {
	return Primitive_List
}

func (*MapStmt) Kind() Primitive {
	return Primitive_Map
}
