package internal

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

	// Keywords
	Token_Import
	Token_Type
	Token_Struct
	Token_Enum
	Token_Union
)

const (
	Bool Primitive = iota
	String
	Uint8
	Uint16
	Uint32
	Uint64
	Int8
	Int16
	Int32
	Int64
	Float32
	Float64
	Binary
	List
	Map
	Type
)

const (
	Struct TypeModifier = iota
	Union
	Enum
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

func (t Token) String() string {
	v, ok := tokenNameMapping[t]
	if !ok {
		return "unknown"
	}

	return v
}
