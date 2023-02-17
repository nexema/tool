package tokenizer

import "errors"

type TokenizerErr struct {
	err       TokenizerErrKind
	arguments []string
}

type TokenizerErrKind error

func (self TokenizerErr) Error() string {
	return self.err.Error()
}

var (
	ErrUnknownToken            TokenizerErrKind = errors.New("unknown token")
	ErrInvalidString           TokenizerErrKind = errors.New("string literals must start and end with a \"")
	ErrInvalidMultilineComment TokenizerErrKind = errors.New("multiline comments must end with */")
)

func NewTokenizerErr(err TokenizerErrKind, args ...string) *TokenizerErr {
	return &TokenizerErr{
		err:       err,
		arguments: args,
	}
}

func (self TokenizerErr) Is(kind TokenizerErrKind) bool {
	return errors.Is(self.err, kind)
}
