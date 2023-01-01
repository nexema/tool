package internal

import (
	"strings"
)

type Token int
type Keyword int
type Primitive int
type TypeModifier int

const (
	Token_Illegal Token = iota
	Token_EOF
	Token_Whitespace

	// Literals
	Token_Ident // fields, struct names
	Token_Keyword

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
	Primitive_Type
	Primitive_List
	Primitive_Map
	Primitive_Null
)

const (
	TypeModifier_Struct TypeModifier = iota
	TypeModifier_Union
	TypeModifier_Enum
)

const (
	Keyword_Type Keyword = iota
	Keyword_Struct
	Keyword_Enum
	Keyword_Union
	Keyword_Import
	Keyword_String
	Keyword_Boolean
	Keyword_Uint
	Keyword_Int
	Keyword_Uint8
	Keyword_Uint16
	Keyword_Uint32
	Keyword_Uint64
	Keyword_Int8
	Keyword_Int16
	Keyword_Int32
	Keyword_Int64
	Keyword_Float32
	Keyword_Float64
	Keyword_List
	Keyword_Map
	Keyword_Binary
	Keyword_Null
	Keyword_As
)

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
}

var inverseKeywordMapping map[string]Keyword = map[string]Keyword{
	"struct":  Keyword_Struct,
	"enum":    Keyword_Enum,
	"union":   Keyword_Union,
	"type":    Keyword_Type,
	"import":  Keyword_Import,
	"string":  Keyword_String,
	"boolean": Keyword_Boolean,
	"uint8":   Keyword_Uint8,
	"uint16":  Keyword_Uint16,
	"uint32":  Keyword_Uint32,
	"uint64":  Keyword_Uint64,
	"int8":    Keyword_Int8,
	"int16":   Keyword_Int16,
	"int32":   Keyword_Int32,
	"int64":   Keyword_Int64,
	"float32": Keyword_Float32,
	"float64": Keyword_Float64,
	"list":    Keyword_List,
	"map":     Keyword_Map,
	"binary":  Keyword_Binary,
	"int":     Keyword_Int,
	"uint":    Keyword_Uint,
	"as":      Keyword_As,
}

var keywordMapping map[Keyword]string = map[Keyword]string{
	Keyword_Struct:  "struct",
	Keyword_Enum:    "enum",
	Keyword_Union:   "union",
	Keyword_Type:    "type",
	Keyword_Import:  "import",
	Keyword_String:  "string",
	Keyword_Boolean: "boolean",
	Keyword_Uint8:   "uint8",
	Keyword_Uint16:  "uint16",
	Keyword_Uint32:  "uint32",
	Keyword_Uint64:  "uint64",
	Keyword_Int8:    "int8",
	Keyword_Int16:   "int16",
	Keyword_Int32:   "int32",
	Keyword_Int64:   "int64",
	Keyword_Float32: "float32",
	Keyword_Float64: "float64",
	Keyword_List:    "list",
	Keyword_Map:     "map",
	Keyword_Binary:  "binary",
	Keyword_Int:     "int",
	Keyword_Uint:    "uint",
	Keyword_Null:    "null",
	Keyword_As:      "as",
}

var primitiveMapping map[string]Primitive = map[string]Primitive{
	"boolean": Primitive_Bool,
	"string":  Primitive_String,
	"uint8":   Primitive_Uint8,
	"uint16":  Primitive_Uint16,
	"uint32":  Primitive_Uint32,
	"uint64":  Primitive_Uint64,
	"int8":    Primitive_Int8,
	"int16":   Primitive_Int16,
	"int32":   Primitive_Int32,
	"int64":   Primitive_Int64,
	"int":     Primitive_Int32,
	"uint":    Primitive_Uint32,
	"float32": Primitive_Float32,
	"float64": Primitive_Float64,
	"binary":  Primitive_Binary,
	"list":    Primitive_List,
	"map":     Primitive_Map,
	"null":    Primitive_Null,
}

func (t Token) String() string {
	v, ok := tokenNameMapping[t]
	if !ok {
		return "unknown"
	}

	return v
}

func (k Keyword) String() string {
	s := keywordMapping[k]
	return s
}

func isKeyword(s string) bool {
	_, ok := inverseKeywordMapping[strings.ToLower(s)]
	return ok
}

func isExactKeyword(s string, keyword Keyword) bool {
	keywordString := keywordMapping[keyword]
	return keywordString == strings.ToLower(s)
}

func parseTypeModifier(s string) (TypeModifier, bool) {
	switch s {
	case "struct":
		return TypeModifier_Struct, true
	case "union":
		return TypeModifier_Union, true
	case "enum":
		return TypeModifier_Enum, true
	default:
		return TypeModifier_Struct, false
	}
}
