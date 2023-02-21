package analyzer

import (
	"errors"
	"fmt"
	"strings"

	"tomasweigenast.com/nexema/tool/definition"
	"tomasweigenast.com/nexema/tool/token"
	"tomasweigenast.com/nexema/tool/tokenizer"
)

type AnalyzerError struct {
	At   tokenizer.Pos
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

type (
	ErrUnknownTypeModifier struct {
		Token token.TokenKind
	}

	ErrNeedAlias struct{}

	ErrTypeNotFound struct {
		Name  string
		Alias string
	}

	ErrNotValidBaseType struct {
		Name  string
		Alias string
	}

	ErrAlreadyDefined struct {
		Name string
	}

	ErrWrongArgumentsLen struct {
		Primitive    definition.ValuePrimitive
		ArgumentsLen int
	}

	ErrWrongArguments struct {
		Primitive definition.ValuePrimitive
		IsMapKey  bool
	}

	ErrWrongFieldIndex struct {
		Err error
	}

	ErrAssignmentKeyAlreadyInUse struct {
		KeyName string
	}

	ErrWrongAnnotationValue struct{}

	ErrIllegalUseCycle struct {
		TypeName string
	}

	ErrNonNullableUnionFields struct{}
)

func (e ErrWrongArgumentsLen) Message() string {
	if e.Primitive == definition.List {
		return fmt.Sprintf("list expects exactly one argument, got %d instead", e.ArgumentsLen)
	} else {
		return fmt.Sprintf("map expects exactly two arguments, got %d instead", e.ArgumentsLen)
	}
}

func (e ErrWrongAnnotationValue) Message() string {
	return "annotation value must be a value of type string, int64, float64 or boolean"
}

func (e ErrNonNullableUnionFields) Message() string {
	return "unions cannot declare nullable fields"
}

func (e ErrIllegalUseCycle) Message() string {
	return fmt.Sprintf("cannot declare a field in %[1]q whose value type is %[1]q", e.TypeName)
}

func (e ErrAssignmentKeyAlreadyInUse) Message() string {
	return fmt.Sprintf("assignment key %q already in use", e.KeyName)
}

func (e ErrWrongFieldIndex) Message() string {
	return e.Err.Error()
}

func (e ErrWrongArguments) Message() string {
	if e.Primitive == definition.List {
		return "list argument cannot be another list or a map"
	} else {
		if e.IsMapKey {
			return "map key must be a non-nullable string, int, uint, fixed-int, fixed-uint or boolean"
		} else {
			return "map value cannot be another list or a map"
		}
	}
}

func (e ErrUnknownTypeModifier) Message() string {
	return fmt.Sprintf("unknown type's modifier %s", e.Token)
}

func (ErrNeedAlias) Message() string {
	return "more than one object is defined with the same name, try aliasing your imports"
}

func (e ErrAlreadyDefined) Message() string {
	return fmt.Sprintf("%q is already defined", e.Name)
}

func (e ErrNotValidBaseType) Message() string {
	fullName := e.Alias
	if len(e.Alias) > 0 {
		fullName += "." + e.Name
	}
	return fmt.Sprintf("%q is not a valid base type", fullName)
}

func (e ErrTypeNotFound) Message() string {
	fullName := e.Alias
	if len(e.Alias) > 0 {
		fullName += "." + e.Name
	}

	return fmt.Sprintf("type %q not found, are you missing an import?", fullName)
}

func NewAnalyzerError(err AnalyzerErrorKind, at tokenizer.Pos) *AnalyzerError {
	return &AnalyzerError{at, err}
}

type AnalyzerErrorCollection []*AnalyzerError

func newAnalyzerErrorCollection() *AnalyzerErrorCollection {
	collection := make(AnalyzerErrorCollection, 0)
	return &collection
}

func (self *AnalyzerErrorCollection) push(kind AnalyzerErrorKind, at tokenizer.Pos) {
	(*self) = append((*self), NewAnalyzerError(kind, at))
}

func (self *AnalyzerErrorCollection) IsEmpty() bool {
	return len(*self) == 0
}

func (self *AnalyzerErrorCollection) Display() string {
	out := make([]string, len(*self))
	for i, err := range *self {
		out[i] = fmt.Sprintf("%d:%d -> %s", err.At.Line, err.At.Start, err.Kind.Message())
	}

	return strings.Join(out, "\n")
}
