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
	metadata     *mapStmt               // type's metadata
	name         string                 // type's name
	typeModifier TypeModifier           // type's modifier (union,struct,enum)
	fields       *fieldsStmt            // type's fields
	docs         *documentationComments // type's documentation comments
}

func (t *typesStmt) add(typestmt *typeStmt) {
	(*t) = append((*t), typestmt)
}

func (i *importsStmt) add(importStmt *importStmt) {
	(*i) = append((*i), importStmt)
}

type fieldsStmt []*fieldStmt
type fieldStmt struct {
	index        int
	name         string
	valueType    *valueTypeStmt
	defaultValue baseIdentifierStmt
	metadata     *mapStmt
	docs         *documentationComments
}

type valueTypeStmt struct {
	nullable       bool
	primitive      Primitive
	typeArguments  *[]*valueTypeStmt
	customTypeName *string // the name of the custom type if primitive is Primitive_Type
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

type customTypeIdentifierStmt struct {
	customTypeName string
	value          string
}

type listStmt []*identifierStmt

func (l *listStmt) add(i *identifierStmt) {
	(*l) = append((*l), i)
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

func (c *customTypeIdentifierStmt) Primitive() Primitive {
	return Primitive_Type
}

type commentStmt struct {
	text        string
	commentType commentType
}

type commentType int8

const (
	singleline commentType = iota
	multiline
)

type documentationComments []*commentStmt

func (d *documentationComments) add(c *commentStmt) {
	d.add(c)
}