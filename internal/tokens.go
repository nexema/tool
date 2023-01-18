package internal

import (
	"regexp"
)

type Token int8
type Primitive int8

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
	Token_Type // type

	modifiers_beg
	Token_Struct // struct
	Token_Enum   // enum
	Token_Union  // union
	modifiers_end
	Token_Import // import
	Token_As
	keywords_end
)

const (
	Primitive_Illegal Primitive = iota
	Primitive_Uint8
	Primitive_Uint16
	Primitive_Uint32
	Primitive_Uint64
	Primitive_Int8
	Primitive_Int16
	Primitive_Int32
	Primitive_Int64
	Primitive_Float32
	Primitive_Float64
	Primitive_String
	Primitive_Bool
	Primitive_Binary
	Primitive_Map
	Primitive_List
	Primitive_Type
	Primitive_Null
)

const nullKeyword = "null"
const importKeyword = "import"

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
	Token_Import:   "import",
	Token_As:       "as",
}

var primitiveMapping map[Primitive]string = map[Primitive]string{
	Primitive_Null:    "null",
	Primitive_Uint8:   "uint8",
	Primitive_Uint16:  "uint16",
	Primitive_Uint32:  "uint32",
	Primitive_Uint64:  "uint64",
	Primitive_Int8:    "int8",
	Primitive_Int16:   "int16",
	Primitive_Int32:   "int32",
	Primitive_Int64:   "int64",
	Primitive_String:  "string",
	Primitive_Binary:  "binary",
	Primitive_Bool:    "bool",
	Primitive_List:    "list",
	Primitive_Map:     "map",
	Primitive_Float32: "float32",
	Primitive_Float64: "float64",
	Primitive_Type:    "type",
}

var keywords map[string]Token
var primitives map[string]Primitive
var identifierRegex = regexp.MustCompile(`[A-Za-z_][A-Za-z_0-9]*`)

func init() {
	keywords = make(map[string]Token, keywords_end-(keywords_beg+1))
	for i := keywords_beg + 1; i < keywords_end; i++ {
		keywords[tokenMapping[i]] = i
	}

	primitives = make(map[string]Primitive)
	for primitive, name := range primitiveMapping {
		primitives[name] = primitive
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

// GetPrimitive returns the Primitive the ident represents
func GetPrimitive(ident string) Primitive {
	prim, ok := primitives[ident]
	if ok {
		return prim
	}

	return Primitive_Illegal
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

func (tok Token) IsModifier() bool {
	return modifiers_beg < tok && tok < modifiers_end
}

// IsIdentifier returns a boolean indicating if the current string is an identifier, that is, a string with
// matches the following regex: [A-Za-z_][A-Za-z_0-9]* and is not a keyword
func IsIdentifier(i string) bool {
	if i == "" || IsKeyword(i) {
		return false
	}

	return identifierRegex.MatchString(i)
}

func IsPrimitive(i string) bool {
	_, ok := primitives[i]
	return ok
}

func (p Primitive) String() string {
	return primitiveMapping[p]
}

func (p Primitive) IsInt() bool {
	switch p {
	case Primitive_Int8, Primitive_Int16, Primitive_Int32, Primitive_Int64, Primitive_Uint8, Primitive_Uint16, Primitive_Uint32, Primitive_Uint64:
		return true
	default:
		return false
	}
}

func (p Primitive) IsFloat() bool {
	switch p {
	case Primitive_Float32, Primitive_Float64:
		return true
	default:
		return false
	}
}
