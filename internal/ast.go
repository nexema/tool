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

func (t *typesStmt) add(typestmt *typeStmt) {
	(*t) = append((*t), typestmt)
}

type fieldsStmt []*fieldStmt
type fieldStmt struct {
	index        int
	name         string
	valueType    *valueTypeStmt
	defaultValue baseIdentifierStmt
	metadata     *mapStmt
}

type valueTypeStmt struct {
	nullable      bool
	primitive     Primitive
	typeArguments *[]*valueTypeStmt
}

func (f *fieldsStmt) add(field *fieldStmt) {
	(*f) = append((*f), field)
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

type baseIdentifierStmt interface {
	Primitive() Primitive
}

type identifierStmt struct {
	value     interface{}
	valueType *valueTypeStmt
}

type listStmt []*identifierStmt

func (l *listStmt) add(i *identifierStmt) {
	(*l) = append((*l), i)
}

func (l *listStmt) isEmpty() bool {
	return len(*l) == 0
}

func (i *identifierStmt) Primitive() Primitive {
	return i.valueType.primitive
}

func (l *listStmt) Primitive() Primitive {
	return Primitive_List
}

func (m *mapStmt) Primitive() Primitive {
	return Primitive_Map
}
