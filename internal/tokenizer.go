package internal

import (
	"bufio"
	"io"
	"regexp"
)

var rules map[regexp.Regexp]Token = map[regexp.Regexp]Token{
	*regexp.MustCompile(`/^\d+/`): Token_Number,
}

type Tokenizer struct {
	reader *bufio.Reader
	pos    ScannerPosition
}

type ScannerPosition struct {
	offset int
	line   int
}

func NewTokenizer(r *bufio.Reader) *Tokenizer {
	return &Tokenizer{
		reader: r,
		pos:    ScannerPosition{offset: 0, line: 1},
	}
}

func (t *Tokenizer) scan() (tok Token, lit string, pos ScannerPosition) {
	ch, _, err := t.reader.Peek()
	if err != nil && err == io.EOF {
		return Token_EOF, "", t.pos
	}

	for rule, mapping := range rules {
		rule.MatchString(string(ch))
	}
}
