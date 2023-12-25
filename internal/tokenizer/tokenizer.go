package tokenizer

import (
	"bufio"
	"io"
	"strings"
	"unicode"

	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
)

const (
	eof      rune = -1
	newline  rune = '\n'
	space    rune = ' '
	tab      rune = '\t'
	carriage rune = '\r'
)

type Tokenizer struct {
	reader      *bufio.Reader
	ch          rune
	currentPos  int
	currentLine int

	// only for debugging
	currentChar  string
	nextChar     string
	previousChar string
}

func NewTokenizer(reader io.Reader) *Tokenizer {
	tokenizer := &Tokenizer{
		reader:      bufio.NewReader(reader),
		ch:          eof,
		currentPos:  -1,
		currentLine: 1,
	}

	tokenizer.next()

	return tokenizer
}

func (self *Tokenizer) Next() (tok *token.Token, pos reference.Pos, err *TokenizerErr) {
	self.skipWhitespace()
	if self.ch == eof {
		return token.Token_EOF, reference.NewPos(), nil
	}

	pos = reference.NewPos(self.currentPos, self.currentPos+1, self.currentLine, self.currentLine)
	var tokenKind token.TokenKind

	switch self.ch {
	case ':':
		tokenKind = token.Colon
	case '=':
		tokenKind = token.Assign
	case '{':
		tokenKind = token.Lbrace
	case '}':
		tokenKind = token.Rbrace
	case '[':
		tokenKind = token.Lbrack
	case ']':
		tokenKind = token.Rbrack
	case '(':
		tokenKind = token.Lparen
	case ')':
		tokenKind = token.Rparen
	case ',':
		tokenKind = token.Comma
	case '?':
		tokenKind = token.QuestionMark
	case '#':
		tokenKind = token.Hash
	default:
		next := self.peek()
		if self.ch == '.' {
			if isNumeric(next) {
				return self.readNumber()
			} else {
				self.next()
				return token.NewToken(token.Period, "."), pos, nil
			}
		}

		if (self.ch == '-' && isNumeric(next)) || isNumeric(self.ch) {
			return self.readNumber()
		}

		if self.ch == '"' {
			tok, pos, err = self.readString()
			self.next()
			return
		}

		if self.ch == '/' && (next == '*' || next == '/') {
			tok, pos, err = self.readComment()
			self.next()
			return
		}

		if isAlphabetic(self.ch) {
			tok, pos, err = self.readIdentifier()
			if err == nil {
				// try to convert to a keyword
				keyword := tok.ToKeyword()
				if keyword != nil {
					return keyword, pos, err
				}
			}
			return
		}

		return nil, reference.NewPos(), NewTokenizerErr(errUnknownToken(self.ch), string(self.ch))
	}

	tok = token.NewToken(tokenKind, string(self.ch))
	self.next()

	return tok, pos, nil
}

func (self *Tokenizer) GetCurrentPosition() reference.Pos {
	return reference.NewPos(self.currentPos, self.currentPos, self.currentLine, self.currentLine)
}

func (self *Tokenizer) readIdentifier() (tok *token.Token, pos reference.Pos, err *TokenizerErr) {
	result := new(strings.Builder)
	result.WriteRune(self.ch)
	startPos := self.currentPos
	for {
		ch := self.next()
		if isAlphanumeric(ch) || ch == '_' {
			result.WriteRune(ch)
			continue
		}

		break
	}

	return token.NewToken(token.Ident, result.String()), self.getPos(startPos), nil
}

func (self *Tokenizer) readComment() (tok *token.Token, pos reference.Pos, err *TokenizerErr) {
	result := new(strings.Builder)
	startPos := self.currentPos
	self.next() // initial / was read
	if self.ch == '/' {
		for {
			ch := self.next()
			if ch == newline || ch == eof {
				pos := self.getPos(startPos)
				if ch == newline {
					self.advanceLine()
				}

				return token.NewToken(token.Comment, result.String()), pos, nil
			}

			result.WriteRune(ch)
		}
	} else {
		// * was read, scan until */ is found
		startLine := self.currentLine
		for {
			ch := self.next()
			if ch == eof {
				return nil, reference.NewPos(), NewTokenizerErr(ErrInvalidMultilineComment)
			}

			if ch == '*' && self.peek() == '/' {
				self.next() // consume last /
				return token.NewToken(token.CommentMultiline, result.String()), reference.NewPos(startPos, self.currentPos+1, startLine, self.currentLine), nil
			}

			if ch == newline {
				self.advanceLine()
			}

			result.WriteRune(ch)
		}
	}
}

func (self *Tokenizer) readNumber() (tok *token.Token, pos reference.Pos, err *TokenizerErr) {
	isDecimal := self.ch == '.'
	result := new(strings.Builder)
	result.WriteRune(self.ch)

	startPos := self.currentPos

	for {
		ch := self.next()
		if ch == '.' {
			if isDecimal {
				break
			}

			result.WriteRune('.')
			isDecimal = true
		} else if isNumeric(ch) {
			result.WriteRune(ch)
		} else {
			break
		}
	}

	pos = self.getPos(startPos)
	if isDecimal {
		return token.NewToken(token.Decimal, result.String()), pos, nil
	} else {
		return token.NewToken(token.Integer, result.String()), pos, nil
	}
}

func (self *Tokenizer) readString() (tok *token.Token, pos reference.Pos, err *TokenizerErr) {
	result := new(strings.Builder)
	startPos := self.currentPos
	for {
		ch := self.next()
		if ch == newline || ch == eof {
			return nil, reference.NewPos(), NewTokenizerErr(ErrInvalidString)
		}

		if ch == '"' {
			break
		}

		// escape "
		if ch == '\\' && self.peek() == '"' {
			result.WriteRune('"')
			self.next()
		} else {
			result.WriteRune(ch)
		}
	}

	pos = self.getPos(startPos)
	pos.End++
	return token.NewToken(token.String, result.String()), pos, nil
}

func (self *Tokenizer) skipWhitespace() {
	for {
		if self.ch == newline {
			self.currentLine++
			self.currentPos = -1
			self.next()
			continue
		}

		if self.ch == space || self.ch == tab || self.ch == carriage {
			self.next()
		} else {
			break
		}
	}
}

func (self *Tokenizer) advanceLine() {
	self.currentLine++
	self.currentPos = -1
}

func (self *Tokenizer) next() rune {
	ch, _, err := self.reader.ReadRune()
	self.currentPos++
	if err == nil {
		self.previousChar = string(self.ch)
		self.currentChar = string(ch)
		self.nextChar = string(self.peek())
		self.ch = ch
	} else {
		self.ch = eof
	}

	return self.ch
}

func (self *Tokenizer) peek() rune {
	buf, err := self.reader.Peek(1)
	if err == nil {
		return rune(buf[0])
	}

	return eof
}

func (self *Tokenizer) getPos(start int) reference.Pos {
	return reference.NewPos(
		start,
		self.currentPos,
		self.currentLine,
		self.currentLine,
	)
}

func isNumeric(c rune) bool {
	switch {
	case c >= '0' && c <= '9':
		return true
	default:
		return c > '\x7f' && unicode.IsNumber(c)
	}
}

func isAlphabetic(ch rune) bool {
	if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' {
		return true
	}
	if ch > '\x7f' && unicode.IsLetter(ch) {
		return true
	}
	return false
}

func isAlphanumeric(ch rune) bool {
	return isNumeric(ch) || isAlphabetic(ch)
}
