package parser

import (
	"errors"
	"fmt"
	"strings"

	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
	"tomasweigenast.com/nexema/tool/internal/tokenizer"
)

type ParserError struct {
	At   reference.Reference
	Kind ParserErrorKind
}

type ParserErrorKind interface {
	Message() string
}

type (
	ErrUnexpectedEOF struct{}

	ErrUnexpectedToken struct {
		Expected token.TokenKind
		Got      token.Token
	}

	ErrTokenizer struct {
		Err tokenizer.TokenizerErr
	}

	ErrExpectedIdentifier struct {
		Got token.Token
	}

	ErrNumberParse struct {
		Wrapped error
		Value   string
	}

	ErrInvalidLiteral struct {
		Got token.Token
	}

	ErrUnexpectedValue struct {
		Expected string
		Got      token.Token
	}

	ErrExpectedDeclaration struct {
		Got token.Token
	}

	ErrExpectedLiteral struct {
		Got token.Token
	}
)

func (ErrUnexpectedEOF) Message() string {
	return "unexpected end of file"
}

func (u ErrUnexpectedToken) Message() string {
	if u.Expected == token.Illegal {
		return fmt.Sprintf("unexpected token %q", u.Got.Literal)
	}

	return fmt.Sprintf("expected %q token, got %q instead", u.Expected, u.Got.Literal)
}

func (u ErrTokenizer) Message() string {
	return u.Err.Error()
}

func (u ErrExpectedIdentifier) Message() string {
	return fmt.Sprintf("expected identifier, got %s instead", u.Got)
}

func (u ErrNumberParse) Message() string {
	return fmt.Sprintf("%s is not a valid number", u.Value)
}

func (u ErrInvalidLiteral) Message() string {
	return fmt.Sprintf("%s is not a valid literal value", u.Got)
}

func (u ErrUnexpectedValue) Message() string {
	return fmt.Sprintf("expected %s, got %s instead", u.Expected, u.Got)
}

func (u ErrExpectedDeclaration) Message() string {
	return fmt.Sprintf("expected declaration, got %s instead", u.Got)
}

func (u ErrExpectedLiteral) Message() string {
	return fmt.Sprintf("expected literal, got %s instead", u.Got)
}

func NewParserErr(err ParserErrorKind, at reference.Reference) *ParserError {
	return &ParserError{at, err}
}

type ParserErrorCollection []*ParserError

func newParserErrorCollection() *ParserErrorCollection {
	collection := make(ParserErrorCollection, 0)
	return &collection
}

func (self *ParserErrorCollection) push(err *ParserError) {
	(*self) = append((*self), err)
}

func (self *ParserErrorCollection) IsEmpty() bool {
	return len(*self) == 0
}

func (self *ParserErrorCollection) Display() string {
	out := make([]string, len(*self))
	for i, err := range *self {
		out[i] = fmt.Sprintf("%s -> %s", err.At, err.Kind.Message())
	}

	return strings.Join(out, "\n")
}

func (self *ParserErrorCollection) AsError() error {
	return errors.New(self.Display())
}

// func (self *ParserErrorCollection) Clone() []ParserError {
// 	clone := make([]ParserError, len(*self))
// 	for i, elem := range *self {
// 		clone[i] = *elem
// 	}

// 	return clone
// }
