package internal

import (
	"regexp"
)

type Token int8

const (
	Token_Illegal Token = iota
	Token_EOF
	Token_Whitespace
	Token_Comment // /-started or *-started

	// Literals
	literals_beg
	Token_Ident  // struct's or field's name
	Token_Int    // 5
	Token_Float  // 5.4
	Token_String // "hello world"
	literals_end

	// Operators
	operators_beg
	Token_Assign   // =
	Token_Nullable // ?
	Token_Lparen   // (
	Token_Lbrack   // [
	Token_Lbrace   // {
	Token_Rparen   // )
	Token_Rbrack   // ]
	Token_Rbrace   // }
	Token_Colon    // :
	Token_Comma    // ,
	Token_Period   // .
	Token_At       // @
	operators_end

	// Keywords
	keywords_beg
	Token_Type   // type
	Token_Struct // struct
	Token_Enum   // enum
	Token_Union  // union
	keywords_end
)

var tokenMapping map[Token]string = map[Token]string{
	Token_EOF:      "eof",
	Token_Comment:  "comment",
	Token_Illegal:  "illegal",
	Token_Int:      "int",
	Token_Float:    "float",
	Token_Ident:    "ident",
	Token_String:   "string",
	Token_Assign:   "=",
	Token_Nullable: "?",
	Token_Lparen:   "(",
	Token_Rparen:   ")",
	Token_Lbrace:   "{",
	Token_Rbrace:   "}",
	Token_Lbrack:   "[",
	Token_Rbrack:   "]",
	Token_Colon:    ":",
	Token_Comma:    ",",
	Token_Period:   ".",
	Token_At:       "@",
	Token_Struct:   "struct",
	Token_Union:    "union",
	Token_Enum:     "enum",
	Token_Type:     "type",
}

var keywords map[string]Token
var identifierRegex = regexp.MustCompile(`[A-Za-z_][A-Za-z_0-9]*`)

func init() {
	keywords = make(map[string]Token, keywords_end-(keywords_beg+1))
	for i := keywords_beg + 1; i < keywords_end; i++ {
		keywords[tokenMapping[i]] = i
	}
}

// String returns the string corresponding to the token tok.
func (tok Token) String() string {
	s := tokenMapping[tok]
	return s
}

// GetKeyword returns the keyword the ident represents, or Token_Ident if its not a keyword
func GetKeyword(ident string) Token {
	tok, ok := keywords[ident]
	if ok {
		return tok
	}

	return Token_Ident
}

func (tok Token) IsLiteral() bool {
	return literals_beg < tok && tok < literals_end
}

func (tok Token) IsOperator() bool {
	return operators_beg < tok && tok < operators_end
}

func (tok Token) IsKeyword() bool {
	return keywords_beg < tok && tok < keywords_end
}

func IsKeyword(s string) bool {
	_, ok := keywords[s]
	return ok
}

// IsIdentifier returns a boolean indicating if the current string is an identifier, that is, a string with
// matches the following regex: [A-Za-z_][A-Za-z_0-9]* and is not a keyword
func IsIdentifier(i string) bool {
	if i == "" || IsKeyword(i) {
		return false
	}

	return identifierRegex.MatchString(i)
}
