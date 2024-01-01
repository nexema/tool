package parser

import (
	"fmt"
	"strings"

	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
)

type ParserError struct {
	At   reference.Pos
	Kind ParserErrorKind
}

type ParserErrorKind interface {
	parser()
	Message() string
}

type (
	UnexpectedTokenErrKind struct {
		Expected token.TokenKind
		Got      token.TokenKind
	}

	UnexpectedTokenExpectManyErrKind struct {
		Expected []token.TokenKind
		Got      token.TokenKind
	}

	DuplicatedMapKey struct {
		KeyLiteral string
	}

	TokenizerErrKind struct {
		Msg string
	}
)

func (e UnexpectedTokenErrKind) parser() {}
func (e UnexpectedTokenErrKind) Message() string {
	return fmt.Sprintf("expected token to be %q but got %s instead", e.Expected, e.Got)
}

func (e UnexpectedTokenExpectManyErrKind) parser() {}
func (e UnexpectedTokenExpectManyErrKind) Message() string {
	values := make([]string, len(e.Expected))
	for i, expect := range e.Expected {
		values[i] = fmt.Sprintf(`"%s"`, expect.String())
	}
	expect := strings.Join(values, " or ")

	return fmt.Sprintf("expected token to be %s but got %s instead", expect, e.Got)
}
func (e TokenizerErrKind) parser() {}
func (e TokenizerErrKind) Message() string {
	return e.Msg
}

func (e DuplicatedMapKey) parser() {}
func (e DuplicatedMapKey) Message() string {
	return fmt.Sprintf("duplicated key %q in map literal", e.KeyLiteral)
}
