package parser

import (
	"fmt"
	"strings"

	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
)

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

// Ast represents the abstract syntax tree of a file
type Ast struct {
	File       reference.File
	Statements []Statement
}

func (stmt Ast) TokenLiteral() string { return "Ast" }
func (Ast) statementNode()            {}

// CommentStatement represents a single line or multi line comment
type CommentStatement struct {
	Token token.Token // the token.Comment or token.CommentMultiline token
}

func (stmt CommentStatement) TokenLiteral() string { return stmt.Token.Literal }
func (CommentStatement) statementNode()            {}

// IncludeStatement identifies a '[include] [literal] (as [identifier])' statement.
type IncludeStatement struct {
	Token token.Token // The token.Include token
	Path  LiteralStatement
	Alias *IdentifierStatement
}

func (stmt IncludeStatement) TokenLiteral() string { return stmt.Token.Literal }
func (IncludeStatement) statementNode()            {}

// LiteralStatement represents a literal value, such as a string, a boolean, a number, a list or a map
type LiteralStatement struct {
	Token token.Token // the token that represents the literal
	Value Literal     // the parsed literal value
}

func (stmt LiteralStatement) TokenLiteral() string { return stmt.Token.Literal }
func (LiteralStatement) statementNode()            {}

// IdentifierStatement represents an identifier in the source code, such as the alias of an import
// or the name of a type.
type IdentifierStatement struct {
	Token token.Token  // the identifier itself
	Alias *token.Token // the alias of the identifier
}

func (stmt IdentifierStatement) TokenLiteral() string {
	if stmt.Alias != nil {
		return fmt.Sprintf("%s.%s", stmt.Alias.Literal, stmt.Token.Literal)
	}

	return stmt.Token.Literal
}
func (IdentifierStatement) statementNode() {}

// TypeStatement represents a type definition
type TypeStatement struct {
	Token    token.Token          // the token.Type token
	Name     IdentifierStatement  // the type's name
	Modifier *IdentifierStatement // the type's modifier (struct, enum, base, union)
	Extends  *ExtendsStatement    // the extends statement
	Body     *BlockStatement      // the body of the type
}

func (stmt TypeStatement) TokenLiteral() string { return stmt.Token.Literal }
func (TypeStatement) statementNode()            {}

// FieldStatement represents a field in a type
type FieldStatement struct {
	Token     token.Token           // the token.Ident token (field's name)
	Index     *LiteralStatement     // the field's index
	ValueType *DeclarationStatement // the field's value type
}

func (stmt FieldStatement) TokenLiteral() string { return stmt.Token.Literal }
func (FieldStatement) statementNode()            {}

// DeclarationStatement represents the declaration of a type, for example: string, list(string?), etc
type DeclarationStatement struct {
	Token      token.Token            // the token.Ident token
	Nullable   bool                   // a flag that indicates if the declaration must be nullable
	Arguments  []DeclarationStatement // a list of arguments of the type
	Identifier *IdentifierStatement   // the type's name
}

func (stmt DeclarationStatement) TokenLiteral() string { return stmt.Token.Literal }
func (DeclarationStatement) statementNode()            {}
func (d *DeclarationStatement) Get() (typeName, alias string) {
	typeName = d.Identifier.Token.Literal
	if d.Identifier.Alias != nil {
		alias = d.Identifier.Alias.Literal
	}
	return
}

// ExtendsStatement represents a "extends [typename]" syntax
type ExtendsStatement struct {
	Token    token.Token         // the token.Extends token
	BaseType IdentifierStatement // The basetype's name
}

func (stmt ExtendsStatement) TokenLiteral() string { return stmt.Token.Literal }
func (ExtendsStatement) statementNode()            {}

// BlockStatement represents a list of statements inside a { }
type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (stmt BlockStatement) TokenLiteral() string { return stmt.Token.Literal }
func (BlockStatement) statementNode()            {}

// DefaultsStatement represents a list of default values for a list of fields
type DefaultsStatement struct {
	Token  token.Token       // the token.Defaults token
	Values *LiteralStatement // the values in the defaults block
}

func (stmt DefaultsStatement) TokenLiteral() string { return stmt.Token.Literal }
func (DefaultsStatement) statementNode()            {}

// AnnotationStatement represents a statement in the form of #key = value
type AnnotationStatement struct {
	Token       token.Token // the token.Hash token
	Assignation *AssignStatement
}

func (stmt AnnotationStatement) TokenLiteral() string { return stmt.Token.Literal }
func (AnnotationStatement) statementNode()            {}

// AssignStatement represents the assignation of a value, it is a statement in the form of
// identifier = literal
type AssignStatement struct {
	Token      token.Token // the token.Assign token
	Identifier *IdentifierStatement
	Value      *LiteralStatement
}

func (stmt AssignStatement) TokenLiteral() string { return stmt.Token.Literal }
func (AssignStatement) statementNode()            {}

// Literal represents a literal value, such as a string, a boolean or a number. It represents
// the value itself, not a statement.
type Literal interface {
	Literal() string
	Value() interface{}
}

type BooleanLiteral struct {
	V bool
}

type IntLiteral struct {
	V int64
}

type FloatLiteral struct {
	V float64
}

type StringLiteral struct {
	V string
}

type ListLiteral []Literal

type MapLiteral []MapEntry

type MapEntry struct {
	Key, Value Literal
}

func (self BooleanLiteral) Literal() string    { return fmt.Sprint(self.V) }
func (self BooleanLiteral) Value() interface{} { return self.V }

func (self IntLiteral) Literal() string    { return fmt.Sprint(self.V) }
func (self IntLiteral) Value() interface{} { return self.V }

func (self FloatLiteral) Literal() string    { return fmt.Sprint(self.V) }
func (self FloatLiteral) Value() interface{} { return self.V }

func (self StringLiteral) Literal() string    { return fmt.Sprintf(`"%s"`, self.V) }
func (self StringLiteral) Value() interface{} { return self.V }

func (self ListLiteral) Literal() string {
	return fmt.Sprintf("[%v]", strings.Join(mapArray(self, func(elem Literal) string { return fmt.Sprint(elem.Literal()) }), ", "))
}
func (self ListLiteral) Value() interface{} {
	return mapArray(self, func(elem Literal) interface{} { return elem.Value() })
}

func (self MapLiteral) Literal() string {
	out := make([]string, len(self))
	for _, v := range self {
		out = append(out, fmt.Sprintf("(%v: %v)", v.Key.Literal(), v.Value.Literal()))
	}

	return fmt.Sprintf("[%s]", strings.Join(out, ", "))
}
func (self MapLiteral) Value() interface{} {
	out := make(map[interface{}]interface{}, len(self))
	for _, v := range self {
		out[v.Key.Value()] = v.Value.Value()
	}

	return out
}

func mapArray[T any, O any](in []T, f func(T) O) []O {
	out := make([]O, len(in))
	for i, elem := range in {
		out[i] = f(elem)
	}

	return out
}
