package internal

import (
	"strings"
)

type Token int
type Primitive int
type TypeModifier int

const (
	Token_Illegal Token = iota
	Token_EOF
	Token_Whitespace

	// Literals
	Token_Ident // fields, struct names

	// Misc
	Token_OpenBrackets
	Token_CloseBrackets
	Token_Colon
	Token_QuestionMark
	Token_At
	Token_OpenParens
	Token_CloseParens
	Token_OpenCurlyBraces
	Token_CloseCurlyBraces
	Token_Comma
	Token_Equals
	Token_String

	// Keywords
	Token_Import
	Token_Type
	Token_Struct
	Token_Enum
	Token_Union
)

const (
	Primitive_Bool Primitive = iota
	Primitive_String
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
	Primitive_Binary
	Primitive_List
	Primitive_Map
	Primitive_Type
)

const (
	TypeModifierStruct TypeModifier = iota
	TypeModifierUnion
	TypeModifierEnum
)

var tokenMapping map[string]Token = map[string]Token{
	"import": Token_Import,
	"type":   Token_Type,
	"struct": Token_Struct,
	"union":  Token_Union,
	"enum":   Token_Enum,
}

var tokenNameMapping map[Token]string = map[Token]string{
	Token_Ident:            "identifier",
	Token_OpenBrackets:     "[",
	Token_CloseBrackets:    "]",
	Token_Colon:            ":",
	Token_QuestionMark:     "?",
	Token_At:               "@",
	Token_OpenParens:       "(",
	Token_CloseParens:      ")",
	Token_OpenCurlyBraces:  "{",
	Token_CloseCurlyBraces: "}",
	Token_Comma:            "comma",
	Token_Equals:           "=",
	Token_Import:           "import",
	Token_Type:             "type",
	Token_Struct:           "struct",
	Token_Enum:             "enum",
	Token_Union:            "union",
}

var keywords map[string]bool = map[string]bool{
	"struct":  true,
	"enum":    true,
	"union":   true,
	"type":    true,
	"import":  true,
	"string":  true,
	"bool":    true,
	"uint8":   true,
	"uint16":  true,
	"uint32":  true,
	"uint64":  true,
	"int8":    true,
	"int16":   true,
	"int32":   true,
	"int64":   true,
	"float32": true,
	"float64": true,
	"list":    true,
	"map":     true,
	"binary":  true,
}

func (t Token) String() string {
	v, ok := tokenNameMapping[t]
	if !ok {
		return "unknown"
	}

	return v
}

func isKeyword(s string) bool {
	_, ok := keywords[strings.ToLower(s)]
	return ok
}

func parseTypeModifier(s string) (TypeModifier, bool) {
	switch s {
	case "struct":
		return TypeModifierStruct, true
	case "union":
		return TypeModifierUnion, true
	case "enum":
		return TypeModifierEnum, true
	default:
		return TypeModifierStruct, false
	}
}
