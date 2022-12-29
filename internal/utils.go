package internal

import (
	"fmt"
)

func (p *Parser) expectedError(expected Token, lit string) error {
	return fmt.Errorf("line: %v:%v -> found %q, expected %q", p.buf.pos.line, p.buf.pos.pos, lit, expected.String())
}

func (p *Parser) expectedRawError(expected, given string) error {
	return fmt.Errorf("line: %v:%v -> found %q, expected %s", p.buf.pos.line, p.buf.pos.pos, given, expected)
}

func (p *Parser) keywordGivenErr(given string) error {
	return fmt.Errorf("line: %v:%v -> cannot use %q because its a keyword", p.buf.pos.line, p.buf.pos.pos, given)
}

func (p *Parser) raw(err string) error {
	return fmt.Errorf("line: %v:%v -> %s", p.buf.pos.line, p.buf.pos.pos, err)
}
