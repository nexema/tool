package internal

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

type Scanner struct {
	r   *bufio.Reader
	pos Position
}

type Position struct {
	line int
	pos  int
}

// NewScanner returns a new instance of Scanner
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r:   bufio.NewReader(r),
		pos: Position{line: 1, pos: 0},
	}
}

func (s *Scanner) Scan() (pos Position, token Token, literal string) {
	for {
		ch, _, err := s.r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return s.pos, Token_EOF, ""
			}
		}

		s.pos.pos++
		switch ch {
		case '\n':
			s.resetPos()

		case ':':
			return s.pos, Token_Colon, string(ch)

		case '{':
			return s.pos, Token_OpenCurlyBraces, string(ch)

		case '}':
			return s.pos, Token_CloseCurlyBraces, string(ch)

		case '@':
			return s.pos, Token_At, string(ch)

		case '[':
			return s.pos, Token_OpenBrackets, string(ch)

		case ']':
			return s.pos, Token_CloseBrackets, string(ch)

		case '(':
			return s.pos, Token_OpenParens, string(ch)

		case ')':
			return s.pos, Token_CloseParens, string(ch)

		case '=':
			return s.pos, Token_Equals, string(ch)

		case ',':
			return s.pos, Token_Comma, string(ch)

		case '?':
			return s.pos, Token_QuestionMark, string(ch)

		default:
			if unicode.IsSpace(ch) {
				continue
			} else if unicode.IsDigit(ch) {
				// backup and let scanInt rescan the beginning of the int
				startPos := s.pos
				s.comeback()
				lit := s.scanInt()
				return startPos, Token_Ident, lit
			} else if unicode.IsLetter(ch) {
				startPos := s.pos
				s.comeback()
				lit := s.scanIdent()
				keyword, ok := tokenMapping[lit]
				if ok {
					return startPos, keyword, lit
				}

				return startPos, Token_Ident, lit
			}
		}
	}
}

func (s *Scanner) scanInt() string {
	var lit string
	for {
		ch, _, err := s.r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return lit
			}
		}

		s.pos.pos++
		if unicode.IsDigit(ch) {
			lit += string(ch)
		} else {
			// scanned something not in the integer
			s.comeback()
			return lit
		}
	}
}

func (s *Scanner) scanIdent() string {
	var buf bytes.Buffer
	for {
		ch, _, err := s.r.ReadRune()
		if err != nil && err == io.EOF {
			break
		}

		s.pos.pos++
		if unicode.IsLetter(ch) {
			buf.WriteRune(ch)
		} else {
			s.comeback()
			break
		}
	}

	return buf.String()
}

func (s *Scanner) comeback() {
	if err := s.r.UnreadRune(); err != nil {
		panic(err)
	}

	s.pos.pos--
}

func (s *Scanner) resetPos() {
	s.pos.line++
	s.pos.pos = 0
}
