package internal

// Ast represents the main abstract syntax tree of a messagepack-schema entry
type Ast struct {
	imports *importsStmt
	types   *typesStmt
}

type importsStmt []*importStmt
type importStmt struct {
	src   string
	alias *string
}

type typesStmt []*typeStmt
type typeStmt struct {
	metadata     *mapStmt     // type's metadata
	name         string       // type's name
	typeModifier TypeModifier // type's modifier (union,struct,enum)
	fields       *fieldsStmt  // type's fields
}

type fieldsStmt []*fieldStmt
type fieldStmt struct {
	index        int
	name         string
	nullable     bool
	primitive    Primitive
	defaultValue *identifierStmt
	metadata     *mapStmt
}

type mapStmt []*mapEntryStmt
type mapEntryStmt struct {
	key   *identifierStmt
	value *identifierStmt
}

func (m *mapStmt) add(e *mapEntryStmt) {
	(*m) = append((*m), e)
}

func (m *mapStmt) isEmpty() bool {
	return len(*m) == 0
}

type identifierStmt struct {
	value     interface{}
	valueType Primitive
}

type listStmt []*identifierStmt

func (l *listStmt) add(i *identifierStmt) {
	(*l) = append((*l), i)
}

func (l *listStmt) isEmpty() bool {
	return len(*l) == 0
}
