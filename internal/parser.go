package internal

import (
	"fmt"
	"io"
	"strconv"
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
		// case Token_Import:
		// 	err := p.parseImport()
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// case Token_Type, Token_At:
		// 	typeStmt, err := p.parseType()
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	p.ast.types.add(typeStmt)

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
func (p *Parser) parseType() (*typeStmt, error) {
	tok, lit := p.scan()

	typeStmt := new(typeStmt)

	// read metadata first
	if tok == Token_At {
		mapStmt, err := p.parseMap()
		if err != nil {
			return nil, err
		}

		typeStmt.metadata = mapStmt
	}

	if isKeyword(lit) {

	}

	// scan type's name
	tok, lit = p.scan()
	if tok != Token_Ident {
		return nil, p.expectedRawError("identifier", lit)
	}

	if isKeyword(lit) {
		return nil, p.keywordGivenErr(lit)
	}

	// read modifier
	tok, lit = p.scan()
	if tok == Token_Ident {

		// infer its a struct
		if tok == Token_OpenCurlyBraces {
			typeStmt.typeModifier = TypeModifierStruct
		} else {
			return nil, p.expectedRawError("type modifier", lit)
		}
	}

	modifier, ok := parseTypeModifier(lit)
	if !ok {
		return nil, p.raw(fmt.Sprintf("unknown type modifier: %q", lit))
	}

	typeStmt.typeModifier = modifier

	// read open curly braces
	tok, lit = p.scan()
	if tok != Token_OpenCurlyBraces {
		return nil, p.expectedError(Token_OpenCurlyBraces, lit)
	}

	// read next token
	tok, lit = p.scan()

	// stop reading struct
	if tok == Token_CloseCurlyBraces {
		return nil, nil
	}

	return nil, nil
}

func (p *Parser) parseList() (*listStmt, error) {
	tok, lit := p.scan() // read [
	if tok != Token_OpenBrackets {
		return nil, p.expectedError(Token_OpenBrackets, lit)
	}

	l := new(listStmt)

	added := false
	for {
		tok, _ = p.scan()
		if tok == Token_CloseBrackets { // if we read ], then stop
			break
		}

		// after a value, we expect a comma
		if tok == Token_Comma && added {
			added = false
			continue
		}

		if added {
			if tok == Token_Ident || tok == Token_String {
				return nil, p.raw("list elements must be comma-separated")
			} else {
				return nil, p.raw(`lists must be closed with "]"`)
			}
		}

		if tok == Token_Ident || tok == Token_String {
			p.unscan()
			identifier, err := p.parseIdentifier()
			if err != nil {
				return nil, err
			}

			// add identifier to stmt
			l.add(identifier)
			added = true
		} else {
			break
		}
	}

	return l, nil
}

func (p *Parser) parseMap() (*mapStmt, error) {
	tok, lit := p.scan() // read [
	if tok != Token_OpenBrackets {
		return nil, p.expectedError(Token_OpenBrackets, lit)
	}

	m := new(mapStmt)

	for {
		tok, lit = p.scan()
		if tok == Token_CloseBrackets { // if we read ], then stop
			break
		}

		// after an entry, we expect a comma
		if !m.isEmpty() {
			if tok != Token_Comma {
				return nil, p.raw("map entries must be comma-separated")
			}

			// consume
			tok, lit = p.scan()
		}

		// read (, expect entry
		if tok == Token_OpenParens {

			entry := new(mapEntryStmt)

			// read until we found )
			for {
				tok, _ = p.scan()
				if tok == Token_CloseParens {
					m.add(entry)
					break
				} else if tok == Token_Ident || tok == Token_String {
					p.unscan()
					identifier, err := p.parseIdentifier()
					if err != nil {
						return nil, err
					}

					if entry.key == nil {
						entry.key = identifier

						// next must be colon
						tok, _ = p.scan()
						if tok != Token_Colon {
							return nil, p.raw("key-value pair must be in the format key:value")
						}
					} else if entry.value == nil {
						entry.value = identifier
					} else {
						return nil, p.raw("invalid map declaration")
					}
				} else {
					return nil, p.raw("invalid map declaration")
				}
			}
		} else {
			return nil, p.expectedError(Token_OpenParens, lit)
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
			value:     lit[1 : len(lit)-1],
			valueType: Primitive_String,
		}, nil
	} else {
		// check if input can be parsed into an int
		var value interface{}
		var primitive Primitive
		var err error
		value, err = strconv.ParseInt(lit, 10, 64)

		if err != nil {
			// its not an int, try with float
			value, err = strconv.ParseFloat(lit, 64)
			if err != nil {
				// its not a float, try with bool
				value, err = strconv.ParseBool(lit)
				if err != nil {
					return nil, p.raw(fmt.Sprintf("unknown primitive %s", lit))
					// idk
				} else {
					primitive = Primitive_Bool
				}
			} else {
				primitive = Primitive_Float64
			}
		} else {
			primitive = Primitive_Int64
		}

		return &identifierStmt{
			value:     value,
			valueType: primitive,
		}, nil
	}

	return nil, nil
}
