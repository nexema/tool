package token

import "fmt"

type TokenKind int8

const (
	Illegal TokenKind = iota
	EOF
	Whitespace
	Comment
	CommentMultiline
	String
	Integer
	Decimal
	List
	Map
	Ident
	Type
	Struct
	Enum
	Union
	Base
	Extends
	Rbrace
	Lbrace
	Rbrack
	Lbrack
	Rparen
	Lparen
	Assign
	Colon
	Use
	As
	Comma
	Period
	QuestionMark
	Hash
	Defaults
)

type Token struct {
	Kind    TokenKind
	Literal string
}

var (
	Token_EOF = &Token{Kind: EOF}
)

var tokenKindMap map[TokenKind]string = map[TokenKind]string{
	Comment:          "comment",
	CommentMultiline: "multiline-comment",
	Whitespace:       "whitespace",
	EOF:              "eof",
	Illegal:          "illegal",
	String:           "string",
	Integer:          "integer",
	Decimal:          "decimal",
	Ident:            "ident",
	List:             "[]",
	Map:              "{}",
	Hash:             "#",
	Rbrace:           "}",
	Lbrace:           "{",
	Rparen:           ")",
	Lparen:           "(",
	Rbrack:           "]",
	Lbrack:           "[",
	Assign:           "=",
	Colon:            ":",
	Use:              "use",
	As:               "as",
	Comma:            ",",
	Period:           ".",
	QuestionMark:     "?",
	Extends:          "extends",
	Defaults:         "defaults",
	Base:             "base",
	Struct:           "struct",
	Union:            "union",
	Enum:             "enum",
	Type:             "type",
}

func NewToken(kind TokenKind, literal ...string) *Token {
	var literalValue string
	if len(literal) == 1 {
		literalValue = literal[0]
	} else {
		var ok bool
		literalValue, ok = tokenKindMap[kind]
		if !ok {
			panic(fmt.Errorf("literal for token kind %v not found", kind))
		}
	}

	return &Token{kind, literalValue}
}

func (self *Token) ToKeyword() *Token {
	var kind TokenKind
	switch self.Literal {
	case "type":
		kind = Type
	case "as":
		kind = As
	case "struct":
		kind = Struct
	case "enum":
		kind = Enum
	case "union":
		kind = Union
	case "base":
		kind = Base
	case "extends":
		kind = Extends
	case "use":
		kind = Use
	case "defaults":
		kind = Defaults
	default:
		return nil
	}

	return &Token{kind, self.Literal}
}

func (self *Token) IsEOF() bool {
	return self.Kind == EOF
}

func (self Token) String() string {
	return fmt.Sprintf("%s(%s)", self.Kind, self.Literal)
}

func (self TokenKind) String() string {
	value, ok := tokenKindMap[self]
	if ok {
		return value
	}

	return fmt.Sprintf("Token(%d)", self)
}
