package internal

import (
	"io"
)

type Parser struct {
	s   *Scanner
	ast *Ast // the abstract syntax tree we are building
	buf struct {
		tok Token    // last read token
		lit string   // last read literal
		pos Position // last read token's position
		n   int      // buffer size (max=1)
	}
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		s: NewScanner(r),
	}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead
func (p *Parser) scan() (tok Token, lit string) {
	// if we have a token on the buffer, return it
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// scan next token
	pos, tok, lit := p.s.Scan()

	// save to buffer
	p.buf.pos, p.buf.tok, p.buf.lit = pos, tok, lit
	return
}

// unscan pushes the previously read token back onto the buffer
func (p *Parser) unscan() {
	p.buf.n = 1
}

// Parse parses the given reader and creates an abstract syntax tree
func (p *Parser) Parse() (*Ast, error) {
	p.ast = &Ast{
		imports: new(importsStmt),
		types:   new(typesStmt),
	}
	for {
		tok, lit := p.scan()

		if tok == Token_EOF {
			break
		}

		// to start, only import or type can be specified
		switch tok {
		case Token_Import:
			err := p.parseImport()
			if err != nil {
				return nil, err
			}

		case Token_Type:
			break

		default:
			return nil, p.expectedRawError(`"type" or "import" keywords`, lit)
		}
	}

	return nil, nil
}

// parseImport parses an import keyword
func (p *Parser) parseImport() error {
	// "import" keyword already read, then read identifier
	tok, lit := p.scan()
	if tok != Token_Ident {
		return p.expectedError(Token_Ident, lit)
	}

	return nil
}
