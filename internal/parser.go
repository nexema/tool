package internal

import (
	"fmt"
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
	pos, tok, lit := p.s.Scan(false)

	// save to buffer
	p.buf.pos, p.buf.tok, p.buf.lit = pos, tok, lit
	return
}

// scanWhitespace returns the next token from the underlying scanner without skipping whitespaces.
// If a token has been unscanned then read that instead
func (p *Parser) scanWhitespace() (tok Token, lit string) {
	// if we have a token on the buffer, return it
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// scan next token
	pos, tok, lit := p.s.Scan(true)

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

		case Token_Type, Token_At:
			err := p.parseType()
			if err != nil {
				return nil, err
			}

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

// parseType parses a type
func (p *Parser) parseType() error {
	currenTok := p.buf.tok

	// read metadata first
	if currenTok == Token_At {
		err := p.parseMetadata()
		if err != nil {
			return err
		}
	}

	// scan type's name
	tok, lit := p.scan()
	if tok != Token_Ident {
		return p.expectedRawError("identifier", lit)
	}

	if isKeyword(lit) {
		return p.keywordGivenErr(lit)
	}

	typeStmt := &typeStmt{
		name: lit,
	}

	// read modifier
	tok, lit = p.scan()
	if tok == Token_Ident {

		// infer its a struct
		if tok == Token_OpenCurlyBraces {
			typeStmt.typeModifier = TypeModifierStruct
		} else {
			return p.expectedRawError("identifier", lit)
		}
	}

	modifier, ok := parseTypeModifier(lit)
	if !ok {
		return p.raw(fmt.Sprintf("unknown type modifier: %q", lit))
	}

	typeStmt.typeModifier = modifier

	// read open curly braces
	tok, lit = p.scan()
	if tok != Token_OpenCurlyBraces {
		return p.expectedError(Token_OpenCurlyBraces, lit)
	}

	// read next token
	tok, lit = p.scan()

	// stop reading struct
	if tok == Token_CloseCurlyBraces {
		return nil
	}

	(*p.ast.types) = append((*p.ast.types), typeStmt)
	return nil
}

func (p *Parser) parseMetadata() error {

	return nil
}

func (p *Parser) parseMap() (*mapStmt, error) {
	tok, lit := p.scan() // @ read, read then [
	if tok != Token_OpenBrackets {
		return nil, p.expectedError(Token_OpenParens, lit)
	}

	m := new(mapStmt)

	for {
		tok, lit = p.scan()
		if tok == Token_CloseBrackets { // if we read ], then stop
			break
		}

		// read (, expect entry
		if tok == Token_OpenParens {

			// read until we found )
			for {
				tok, lit = p.scan()
				if tok == Token_CloseParens {
					break
				} else if tok == Token_Ident {
					p.unscan()
					identifier, err := p.parseIdentifier()
					_ = identifier
					_ = err
				}
			}
		}
	}

	return m, nil
}

func (p *Parser) parseIdentifier() (*identifierStmt, error) {
	tok, lit := p.scan()

	// maybe a string
	if tok == Token_String {
		if len(lit) <= 0 {
			return nil, p.raw("invalid string format")
		}

		if lit[len(lit)-1] != '"' {
			return nil, p.raw(`strings must end with quotes (")`)
		}

		return &identifierStmt{
			value:     lit[:len(lit)-1],
			valueType: Primitive_String,
		}, nil
	}

	return nil, nil
}
