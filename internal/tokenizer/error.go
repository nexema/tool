package tokenizer

import (
	"errors"
	"fmt"
)

type TokenizerErr struct {
	err       TokenizerErrKind
	arguments []string
}

type TokenizerErrKind error

func (self TokenizerErr) Error() string {
	if self.err == nil {
		return "unknown"
	}
	return self.err.Error()
}

var (
	ErrInvalidString           TokenizerErrKind = errors.New("string literals must start and end with a \"")
	ErrInvalidMultilineComment TokenizerErrKind = errors.New("multiline comments must end with */")
)

func NewTokenizerErr(err TokenizerErrKind, args ...string) *TokenizerErr {
	return &TokenizerErr{
		err:       err,
		arguments: args,
	}
}

func errUnknownToken(token rune) error {
	return fmt.Errorf("unexpected token %q", token)
}
