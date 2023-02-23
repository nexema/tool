package linker

import (
	"errors"
	"fmt"
	"strings"

	"tomasweigenast.com/nexema/tool/parser"
	"tomasweigenast.com/nexema/tool/tokenizer"
)

type LinkerError struct {
	At   tokenizer.Pos
	Kind LinkerErrorKind
}

type LinkerErrorKind interface {
	Message() string
}

type (
	ErrAlreadyDefined struct {
		Name string
	}

	ErrSelfImport struct{}

	ErrPackageNotFound struct {
		Name string
	}

	ErrCircularDependency struct {
		Src  *parser.File
		Dest *parser.File
	}

	ErrAliasAlreadyDefined struct {
		Alias string
	}
)

func (e ErrAlreadyDefined) Message() string {
	return fmt.Sprintf("%s already defined", e.Name)
}

func (ErrSelfImport) Message() string {
	return "package imported by itself"
}

func (e ErrPackageNotFound) Message() string {
	return fmt.Sprintf("package %q not found", e.Name)
}

func (e ErrCircularDependency) Message() string {
	return fmt.Sprintf("%q imports %q package which generates a not allowed circular dependency", e.Src.Path, e.Dest.Path)
}

func (e ErrAliasAlreadyDefined) Message() string {
	return fmt.Sprintf("alias %q already defined", e.Alias)
}

func NewLinkerErr(err LinkerErrorKind, at tokenizer.Pos) *LinkerError {
	return &LinkerError{at, err}
}

type LinkerErrorCollection []*LinkerError

func newLinkerErrorCollection() *LinkerErrorCollection {
	collection := make(LinkerErrorCollection, 0)
	return &collection
}

func (self *LinkerErrorCollection) push(err *LinkerError) {
	(*self) = append((*self), err)
}

func (self *LinkerErrorCollection) IsEmpty() bool {
	return len(*self) == 0
}

func (self *LinkerErrorCollection) Display() string {
	out := make([]string, len(*self))
	for i, err := range *self {
		out[i] = fmt.Sprintf("%d:%d -> %s", err.At.Line, err.At.Start, err.Kind.Message())
	}

	return strings.Join(out, "\n")
}

func (self *LinkerErrorCollection) AsError() error {
	return errors.New(self.Display())
}
