package token

type TokenKind int8

const (
	Illegal TokenKind = iota
	EOF
	Whitespace
	Comment
	String
	Integer
	Decimal
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

func NewToken(kind TokenKind, literal string) *Token {
	return &Token{kind, literal}
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
