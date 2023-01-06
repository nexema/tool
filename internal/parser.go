package internal

import (
	"bytes"
	"fmt"
	"strconv"
)

// Parser takes tokens from a Tokenizer and produces a Ast
type Parser struct {
	tokenizer *Tokenizer

	tok Token    // current token
	pos Position // current token's position
	lit string   // current token's literal

	comments *[]*CommentStmt
	imports  *[]*ImportStmt
	types    *[]*TypeStmt
}

// NewParser builds a new Parser
func NewParser(buf *bytes.Buffer) *Parser {
	tokenizer := NewTokenizer(buf)
	return &Parser{
		tokenizer: tokenizer,
		pos:       tokenizer.pos,
		tok:       Token_Illegal,
		comments:  new([]*CommentStmt),
		imports:   new([]*ImportStmt),
		types:     new([]*TypeStmt),
	}
}

// Parse parses the given input when creating the Parser and returns
// the corresponding Ast
func (p *Parser) Parse() (*Ast, error) {
	p.next() // read initial token

	// scan any import
	if p.tok == Token_Import && p.nextIs(Token_Colon) {
		// comeback one pos to get into :
		p.undo()
		err := p.parseImportGroup()
		if err != nil {
			return nil, err
		}
	}

	// scan types
	for p.tok == Token_At || p.tok == Token_Type {
		typeStmt, err := p.parseType()
		if err != nil {
			return nil, nil
		}

		(*p.types) = append((*p.types), typeStmt)
		p.next()
	}

	return &Ast{}, nil
}

func (p *Parser) parseImportGroup() error {
	err := p.requireSeq(Token_Import, Token_Colon)
	if err != nil {
		return err
	}

	for p.tok == Token_String {
		importStmt, err := p.parseImport()
		if err != nil {
			return err
		}

		(*p.imports) = append((*p.imports), importStmt)
	}

	return nil
}

// parseImport parses a string and returns it as a ImportStmt
func (p *Parser) parseImport() (*ImportStmt, error) {
	err := p.require(Token_String)
	if err != nil {
		return nil, err
	}

	stmt := &ImportStmt{
		path: &IdentifierStmt{lit: p.lit},
	}
	p.next()

	// may parse AS alias
	if p.tok == Token_As {
		p.next()
		alias, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		stmt.alias = alias
	}

	return stmt, nil
}

// parseType parses a TypeStmt
func (p *Parser) parseType() (*TypeStmt, error) {
	if p.tok == Token_At { // read metadata first

	}

	if p.tok == Token_Type {
		stmt := &TypeStmt{}

		// read name
		ident, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		stmt.Name = ident
		p.next()

		// if read {, then modifier is struct
		if p.tok == Token_Ident {
			// read modifier
			if !p.tok.IsModifier() {
				return nil, p.expectedGiven("expected type modifier")
			}

			err := p.requireNext(Token_Lbrace)
			if err != nil {
				return nil, err
			}
		} else {
			err := p.requireNext(Token_Lbrace)
			if err != nil {
				return nil, err
			}
			stmt.Modifier = Token_Struct
		}

		// read fields until "}"
		for {
			if p.tok == Token_Rbrace {
				goto exit
			}

			if p.tok == Token_EOF {
				break
			}
		}
	}

	return nil, p.require(Token_Rbrace)

exit:
	return nil, nil
}

// parseIdentifier parses an IdentifierStmt. It can be string,
// bool, or any other primitive, enum's values, etc.
func (p *Parser) parseIdentifier() (*IdentifierStmt, error) {
	if p.tok == Token_Ident {
		ident := new(IdentifierStmt)
		lit := p.lit
		p.next()

		// while found a period, it can be import alias, type's name or type's value
		if p.tok == Token_Period {
			p.next()
			err := p.require(Token_Ident)
			if err != nil {
				return nil, err
			}

			ident.alias = lit
			ident.lit = p.lit
		} else {
			ident.lit = lit
		}

		return ident, nil
	} else {
		return nil, p.expectedErr(Token_Ident)
	}
}

func (p *Parser) parseMap() (*MapStmt, error) {
	err := p.requireNext(Token_Lbrack)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// parseList parses an expression in the form: [value1, value2, value3]
func (p *Parser) parseList() (*ListStmt, error) {
	err := p.requireNext(Token_Lbrack)
	if err != nil {
		return nil, err
	}

	list := new(ListStmt)
	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	list.add(value)

	p.next()
	for p.tok == Token_Comma {
		p.next()
		value, err = p.parseValue()
		if err != nil {
			return nil, err
		}
		list.add(value)
		p.next()
	}

	err = p.require(Token_Rbrack)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (p *Parser) parseValue() (ValueStmt, error) {
	if p.tok == Token_String {
		return &PrimitiveValueStmt{kind: Primitive_String, value: p.lit}, nil
	} else if p.tok == Token_Ident {
		kind, value := p.stringToValue()
		if kind != Primitive_Illegal {
			return &PrimitiveValueStmt{kind: kind, value: value}, nil
		}

		// if its illegal, try with enum, it needs the name, and a value.
		enumName, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		// next to scan the value
		p.next()

		// if the next token is a period, we found a declaration in the form
		// alias.EnumName.enum_value, otherwise, enumName contains the EnumName.enum_value
		if p.tok == Token_Period {
			p.next()
			enumValue, err := p.parseIdentifier()
			if err != nil {
				return nil, err
			}

			return &TypeValueStmt{
				typeName: enumName,
				value:    enumValue,
			}, nil
		} else {
			// alias becomes the name, and lit the value
			return &TypeValueStmt{
				typeName: &IdentifierStmt{lit: enumName.alias},
				value:    &IdentifierStmt{lit: enumName.lit},
			}, nil
		}

	} else if p.tok == Token_Float {
		f, _ := strconv.ParseFloat(p.lit, 64)
		return &PrimitiveValueStmt{value: f, kind: Primitive_Float64}, nil
	} else if p.tok == Token_Int {
		i, _ := strconv.ParseInt(p.lit, 10, 64)
		return &PrimitiveValueStmt{value: i, kind: Primitive_Int64}, nil
	} else {
		return nil, p.expectedGiven("string, int, float or identifier")
	}
}

// next reads the next token from the underlying tokenizer.
// Any comment that is encountered, is saved for later use
func (p *Parser) next() error {
	err := p.consume()

	if err != nil {
		return err
	}

	// save any comment
	if p.tok == Token_Comment {
		comment, err := p.parseComment()
		if err != nil {
			return err
		}

		(*p.comments) = append((*p.comments), comment)
	}

	return nil
}

// nextIs advances one token and returns true if its equal to tok
func (p *Parser) nextIs(tok Token) bool {
	p.next()
	defer p.next()
	return p.tok == tok
}

func (p *Parser) undo() {
	// unscan 2 because if we unscan 1 when we "next", it will stay in the same token
	p.tokenizer.unscan(2)
	p.next()
}

// consume reads a token from the underlying tokenizer and saves it in the Parser
func (p *Parser) consume() error {
	var err error
	p.pos, p.tok, p.lit, err = p.tokenizer.Scan()
	return err
}

// parseComment parses the next comment
func (p *Parser) parseComment() (*CommentStmt, error) {
	// scan for /*-style comments to report line end
	lineEnd := p.pos.line
	if p.lit[1] == '*' {
		for _, ch := range p.lit {
			if ch == '\n' {
				lineEnd++
			}
		}
	}

	return &CommentStmt{
		text:      p.lit,
		posStart:  p.pos.offset,
		posEnd:    p.pos.offset + len(p.lit),
		lineStart: p.pos.line,
		lineEnd:   lineEnd,
	}, nil
}

// require returns an error if the next token does not match tok
func (p *Parser) require(tok Token) error {

	// try to move one position
	if p.tok == Token_Illegal {
		p.next()
	}

	if p.tok != tok {
		return p.expectedErr(tok)
	}
	return nil
}

// requireNext does the same as require but moves to the next token after matching
func (p *Parser) requireNext(tok Token) error {

	// try to move one position
	if p.tok == Token_Illegal {
		p.next()
	}

	if p.tok != tok {
		return p.expectedErr(tok)
	}

	p.next()
	return nil
}

// requireMove does the same as require but tries to move one position trying to fetch the required token
func (p *Parser) requireMove(tok Token) error {

	// try to move one position
	if p.tok == Token_Illegal {
		p.next()
	}

	if p.tok != tok {
		p.next()

		if p.tok != tok {
			return p.expectedErr(tok)
		}
	}
	return nil
}

// requireSeq checks if the next sequence of two tokens matches first and second
func (p *Parser) requireSeq(first Token, second Token) error {
	// try to move to a legal position
	if p.tok == Token_Illegal {
		p.next()
	}

	if p.tok != first {
		return p.expectedErr(first)
	}

	p.next()
	if p.tok != second {
		return p.expectedErr(second)
	}

	p.next()
	return nil
}

func (p *Parser) expectedErr(tok Token) error {
	return fmt.Errorf("%s -> expected %q, given %q (%s)", p.pos.String(), tok.String(), p.tok.String(), p.lit)
}

func (p *Parser) err(txt string) error {
	return fmt.Errorf("%s -> %s", p.pos.String(), txt)
}

func (p *Parser) expectedGiven(txt string) error {
	return fmt.Errorf("%s -> %s, given %q", p.pos.String(), txt, p.lit)
}

func (p *Parser) stringToValue() (primitive Primitive, value interface{}) {
	lit := p.lit

	if lit == null {
		return Primitive_Null, nil
	}

	// try with bool
	b, err := strconv.ParseBool(lit)
	if err == nil {
		return Primitive_Bool, b
	}

	return Primitive_Illegal, nil
}
