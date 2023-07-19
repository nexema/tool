package utils

import (
	"fmt"
	"reflect"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/token"
)

type typeBuilder struct {
	typeStmt *parser.TypeStmt
}

type fieldBuilder struct {
	fieldStmt *parser.FieldStmt
}

type valueTypeBuilder struct {
	declStmt *parser.DeclStmt
}

func NewTypeBuilder(name string) *typeBuilder {
	return &typeBuilder{
		typeStmt: &parser.TypeStmt{
			Name: *NewIdentStmt(name),
		},
	}
}

func (self *typeBuilder) Modifier(modifier token.TokenKind) *typeBuilder {
	self.typeStmt.Modifier = modifier
	return self
}

func (self *typeBuilder) Base(typeName string) *typeBuilder {
	self.typeStmt.BaseType = NewDeclStmt(typeName, "", nil, false)
	return self
}

func (self *typeBuilder) Field(field *parser.FieldStmt) *typeBuilder {
	if self.typeStmt.Fields == nil {
		self.typeStmt.Fields = make([]parser.FieldStmt, 0)
	}

	self.typeStmt.Fields = append(self.typeStmt.Fields, *field)
	return self
}

func (self *typeBuilder) Default(key string, value any) *typeBuilder {
	if self.typeStmt.Defaults == nil {
		self.typeStmt.Defaults = make([]parser.AssignStmt, 0)
	}

	self.typeStmt.Defaults = append(self.typeStmt.Defaults, NewAssignStmt(key, value))
	return self
}

func (self *typeBuilder) Result() *parser.TypeStmt {
	return self.typeStmt
}

func NewFieldBuilder(name string) *fieldBuilder {
	return &fieldBuilder{fieldStmt: &parser.FieldStmt{Name: *NewIdentStmt(name)}}
}

func (self *fieldBuilder) Index(number int) *fieldBuilder {
	self.fieldStmt.Index = NewIdentStmt(fmt.Sprint(number))
	return self
}

func (self *fieldBuilder) ValueType(valueType *parser.DeclStmt) *fieldBuilder {
	self.fieldStmt.ValueType = valueType
	return self
}

func (self *fieldBuilder) BasicValueType(valueType string, nullable bool) *fieldBuilder {
	self.fieldStmt.ValueType = NewDeclStmt(valueType, "", nil, nullable)
	return self
}

func (self *fieldBuilder) Result() *parser.FieldStmt {
	return self.fieldStmt
}

func NewValueTypeBuilder(valueType string) *valueTypeBuilder {
	return &valueTypeBuilder{declStmt: &parser.DeclStmt{Token: *token.NewToken(token.Ident, valueType)}}
}

func (self *valueTypeBuilder) Alias(value string) *valueTypeBuilder {
	self.declStmt.Alias = NewIdentStmt(value)
	return self
}

func (self *valueTypeBuilder) Nullable() *valueTypeBuilder {
	self.declStmt.Nullable = true
	return self
}

func (self *valueTypeBuilder) Args(args ...parser.DeclStmt) *valueTypeBuilder {
	self.declStmt.Args = args
	return self
}

func (self *valueTypeBuilder) Result() *parser.DeclStmt {
	return self.declStmt
}

func NewFieldStmt(name string, index int, valueType *parser.DeclStmt) parser.FieldStmt {
	return parser.FieldStmt{
		Name:      *NewIdentStmt(name),
		Index:     NewIdentStmt(fmt.Sprint(index)),
		ValueType: valueType,
	}
}

func NewDeclStmt(value string, alias string, args []string, nullable bool) *parser.DeclStmt {
	var arguments []parser.DeclStmt
	if args != nil {
		arguments = make([]parser.DeclStmt, len(args))
		for i, arg := range args {
			arguments[i] = *NewDeclStmt(arg, "", nil, false)
		}
	}

	return &parser.DeclStmt{
		Token:    *token.NewToken(token.Ident, value),
		Args:     arguments,
		Nullable: nullable,
		Alias:    NewIdentStmt(alias),
	}
}

func NewSimpleDeclStmt(value string, nullable bool) *parser.DeclStmt {

	return &parser.DeclStmt{
		Token:    *token.NewToken(token.Ident, value),
		Nullable: nullable,
	}
}

func NewFullDeclStmt(value string, alias string, args []parser.DeclStmt, nullable bool) *parser.DeclStmt {
	var arguments []parser.DeclStmt
	if args != nil {
		arguments = make([]parser.DeclStmt, len(args))
		for i, arg := range args {
			arguments[i] = arg
		}
	}

	return &parser.DeclStmt{
		Token:    *token.NewToken(token.Ident, value),
		Args:     arguments,
		Nullable: nullable,
		Alias:    NewIdentStmt(alias),
	}
}

func NewTypeStmt(name string, modifier token.TokenKind, fields []parser.FieldStmt, defaults map[string]any) *parser.TypeStmt {
	var defaultsToken []parser.AssignStmt
	if defaults != nil {
		defaultsToken = make([]parser.AssignStmt, len(defaults))
		idx := 0
		for key, value := range defaults {
			defaultsToken[idx] = NewAssignStmt(key, value)
			idx++
		}
	}

	return &parser.TypeStmt{
		Name:     parser.IdentStmt{Token: *token.NewToken(token.String, name)},
		Modifier: modifier,
		Fields:   fields,
		Defaults: defaultsToken,
	}
}

func NewAssignStmt(key string, value any) parser.AssignStmt {
	return parser.AssignStmt{
		Token: *token.NewToken(token.Assign),
		Left:  *NewIdentStmt(key),
		Right: NewLiteralStmt(value),
	}
}

func NewIdentStmt(value string) *parser.IdentStmt {
	return &parser.IdentStmt{Token: *token.NewToken(token.Ident, value)}
}

func NewLiteralStmt(value any) parser.LiteralStmt {
	switch k := value.(type) {
	case string:
		return parser.LiteralStmt{Token: *token.NewToken(token.String, k), Kind: parser.MakeStringLiteral(k)}

	case bool:
		return parser.LiteralStmt{Token: *token.NewToken(token.Ident, fmt.Sprint(value)), Kind: parser.MakeBooleanLiteral(k)}

	case int64:
		return parser.LiteralStmt{Token: *token.NewToken(token.Integer, fmt.Sprint(value)), Kind: parser.MakeIntLiteral(k)}

	case float64:
		return parser.LiteralStmt{Token: *token.NewToken(token.Decimal, fmt.Sprint(value)), Kind: parser.MakeFloatLiteral(k)}
	}

	t := reflect.ValueOf(value)
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		values := make([]parser.LiteralStmt, t.Len())
		for i := 0; i < t.Len(); i++ {
			values[i] = NewLiteralStmt(t.Index(i).Interface())
		}

		return parser.LiteralStmt{Token: *token.NewToken(token.Integer, fmt.Sprint(value)), Kind: parser.MakeListLiteral(values...)}
	} else if t.Kind() == reflect.Map {
		values := make([]parser.MapEntry, t.Len())
		for i, key := range t.MapKeys() {
			value := t.MapIndex(key)
			values[i] = parser.MapEntry{Key: NewLiteralStmt(key), Value: NewLiteralStmt(value)}
		}

		return parser.LiteralStmt{Token: *token.NewToken(token.Integer, fmt.Sprint(value)), Kind: parser.MakeMapLiteral(values...)}
	} else {
		panic(fmt.Sprintf("unable to handle value %v when creating a LiteralStmt", value))
	}
}

func NewUseStmt(path, alias string) *parser.UseStmt {
	stmt := &parser.UseStmt{
		Token: *token.NewToken(token.Use, "use"),
		Path: parser.LiteralStmt{
			Token: *token.NewToken(token.String),
			Kind:  parser.MakeStringLiteral(path),
		},
	}

	if len(alias) > 0 {
		stmt.Alias = NewIdentStmt(alias)
	}

	return stmt
}
