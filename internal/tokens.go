package internal

type Token int8

const (
	Token_Illegal Token = iota
	Token_EOF
	Token_Number
	Token_String
)
