package internal

import (
	"errors"
	"fmt"
	"strings"
)

type ErrorCollection []error

func NewErrorCollection() *ErrorCollection {
	return &ErrorCollection{}
}

func (e *ErrorCollection) Report(err error) {
	(*e) = append((*e), err)
}

func (e *ErrorCollection) IsEmpty() bool {
	return len(*e) == 0
}

func (e *ErrorCollection) Format() error {
	buf := new(strings.Builder)
	for _, err := range *e {
		buf.WriteString(fmt.Sprintf("- %s \n", err.Error()))
	}

	return errors.New(buf.String())
}
