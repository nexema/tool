package analyzer

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
)

type AnalyzerError struct {
	At   reference.Pos
	Kind AnalyzerErrorKind
}

type AnalyzerErrorKind interface {
	Message() string
}

var (
	ErrBaseWrongFieldIndex_EnumShouldBeZeroBased  error = errors.New("first enum's field's index should be 0")
	ErrBaseWrongFieldIndex_EnumShouldBeSubsequent error = errors.New("enum field's should be subsequent")
	ErrBaseWrongFieldIndex_DuplicatedIndex        error = errors.New("fields index must not be duplicated")
)

// common analyzers
type (
	ErrUnknownTypeModifier struct {
		Token token.TokenKind
	}

	// ErrNeedAlias is an error which indicates an import needs an alias, because of duplicated objects
	ErrNeedAlias struct{}

	// ErrTypeNotFound indicates that an object was not found
	ErrTypeNotFound struct {
		Name  string
		Alias string
	}

	ErrAlreadyDefined struct {
		Name string
	}
)

func (e ErrUnknownTypeModifier) Message() string {
	return fmt.Sprintf("unknown type's modifier %s", e.Token)
}

func (ErrNeedAlias) Message() string {
	return "more than one object is defined with the same name, try aliasing your imports"
}

func (e ErrAlreadyDefined) Message() string {
	return fmt.Sprintf("%q is already defined", e.Name)
}

func (e ErrTypeNotFound) Message() string {
	fullName := e.Alias
	if len(e.Alias) > 0 {
		fullName += "." + e.Name
	} else {
		fullName += e.Name
	}

	return fmt.Sprintf("type %q not found, are you missing an import?", fullName)
}

func NewAnalyzerError(err AnalyzerErrorKind, at reference.Pos) *AnalyzerError {
	return &AnalyzerError{at, err}
}

type AnalyzerErrorCollection []*AnalyzerError

func NewAnalyzerErrorCollection() *AnalyzerErrorCollection {
	collection := make(AnalyzerErrorCollection, 0)
	return &collection
}

var errTypeNotFoundKind = reflect.TypeOf(ErrTypeNotFound{})

func (self *AnalyzerErrorCollection) Push(kind AnalyzerErrorKind, at reference.Pos) {
	(*self) = append((*self), NewAnalyzerError(kind, at))
}

func (self *AnalyzerErrorCollection) IsEmpty() bool {
	return len(*self) == 0
}

func (self *AnalyzerErrorCollection) Iterate(f func(err *AnalyzerError)) {
	for _, err := range *self {
		f(err)
	}
}

func (self *AnalyzerErrorCollection) Display() string {
	out := make(map[string]bool, len(*self))
	for _, err := range *self {
		format := fmt.Sprintf("%d:%d -> %s", err.At.Line, err.At.Start, err.Kind.Message())
		if ok := out[format]; ok {
			continue
		}

		out[format] = true
	}

	values := make([]string, len(out))
	i := 0
	for key := range out {
		values[i] = key
		i++
	}

	return strings.Join(values, "\n")
}

func (self *AnalyzerErrorCollection) AsError() error {
	return errors.New(self.Display())
}
