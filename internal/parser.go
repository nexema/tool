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
	if p.tok == Token_Import && p.nextIs(Token_Colon, false) {
		// comeback one pos to get into :
		p.undo(len(importKeyword))
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

	return &Ast{
		imports: p.imports,
	}, nil
}

// parseImportGroup parses a expression in the form import:\n "import1"\n "import2"
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

// parseField parses a FieldStmt
func (p *Parser) parseField() (*FieldStmt, error) {
	stmt := new(FieldStmt)

	// maybe read field index
	if p.tok == Token_Int {
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		stmt.index = value
		p.next()
	}

	// field's name
	if p.tok == Token_Ident {
		value, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		stmt.name = value
	} else {
		return nil, p.expectedErr(Token_Ident)
	}

	// read ":" (required)
	err := p.requireMove(Token_Colon)
	if err != nil {
		return nil, err
	}

	// read value type
	p.next()
	if p.tok == Token_Ident {
		value, err := p.parseValueTypeStmt()
		if err != nil {
			return nil, err
		}

		stmt.valueType = value
	} else {
		return nil, p.expectedErr(Token_Ident)
	}

	if p.tok == Token_Assign { // default value
		p.next()
		value, err := p.parseGenericValue()
		if err != nil {
			return nil, err
		}

		stmt.defaultValue = value
		p.next()
	}

	if p.tok == Token_At { // metadata
		p.next()
		value, err := p.parseMap()
		if err != nil {
			return nil, err
		}

		stmt.metadata = value
		p.next()
	}

	return stmt, nil
}

// parseEnumField parses a FieldStmt but using Enum's syntax.
// Enum's fields uses the same signature as Struct or Union's fields but
// type is not allowed nor default value
func (p *Parser) parseEnumField() (*FieldStmt, error) {
	stmt := new(FieldStmt)

	// maybe read field index
	if p.tok == Token_Int {
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		stmt.index = value
		p.next()
	}

	// field's name
	if p.tok == Token_Ident {
		value, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		stmt.name = value
	} else {
		return nil, p.expectedErr(Token_Ident)
	}

	if p.tok == Token_At { // metadata
		p.next()
		value, err := p.parseMap()
		if err != nil {
			return nil, err
		}

		stmt.metadata = value
		p.next()
	}

	return stmt, nil
}

// parseType parses a TypeStmt
func (p *Parser) parseType() (*TypeStmt, error) {
	stmt := new(TypeStmt)

	// if until this moment any comment has been read, add as type's documentation
	if len(*p.comments) > 0 {
		stmt.documentation = p.comments
		p.comments = nil // clear for new comments
	}

	if p.tok == Token_At { // metadata present
		p.next()
		// parse map
		mapStmt, err := p.parseMap()
		if err != nil {
			return nil, err
		}

		stmt.metadata = mapStmt
		p.next()
	}

	if p.tok != Token_Type {
		return nil, p.expectedErr(Token_Type)
	}
	p.next()

	// read name
	ident, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	stmt.name = ident
	// p.next()

	// if read {, then modifier is struct
	if p.tok == Token_Lbrace {
		stmt.modifier = Token_Struct
	} else {
		// read modifier
		if !p.tok.IsModifier() {
			return nil, p.expectedGiven("expected type modifier")
		}

		stmt.modifier = p.tok

		p.next()
		err := p.require(Token_Lbrace)
		if err != nil {
			return nil, err
		}
	}
	// read fields until "}"
	p.next()
	for {
		if p.tok == Token_Rbrace {
			goto exit
		}

		if p.tok == Token_EOF {
			break
		}

		if stmt.fields == nil {
			stmt.fields = new([]*FieldStmt)
		}

		var fieldStmt *FieldStmt
		var err error
		if stmt.modifier == Token_Enum {
			fieldStmt, err = p.parseEnumField()
		} else {
			fieldStmt, err = p.parseField()
		}

		if err != nil {
			return nil, err
		}

		(*stmt.fields) = append((*stmt.fields), fieldStmt)
	}

	return nil, p.require(Token_Rbrace)

exit:
	return stmt, nil
}

// parseValueTypeStmt parses an ValueTypeStmt.
func (p *Parser) parseValueTypeStmt() (*ValueTypeStmt, error) {
	stmt := new(ValueTypeStmt)
	var err error

	// first parse type's name
	stmt.ident, err = p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	// if we read parens, it contains type arguments
	if p.tok == Token_Lparen {
		p.next()
		stmt.typeArguments = new([]*ValueTypeStmt)

		// read first one
		valueType, err := p.parseValueTypeStmt()
		if err != nil {
			return nil, err
		}

		(*stmt.typeArguments) = append((*stmt.typeArguments), valueType)

		for p.tok == Token_Comma {
			p.next()
			valueType, err := p.parseValueTypeStmt()
			if err != nil {
				return nil, err
			}

			(*stmt.typeArguments) = append((*stmt.typeArguments), valueType)
		}

		// must read )
		err = p.require(Token_Rparen)
		if err != nil {
			return nil, err
		}

		p.next()
	}

	// nullable
	if p.tok == Token_Nullable {
		stmt.nullable = true
		p.next()
	}

	return stmt, err
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
			p.next()
		} else {
			ident.lit = lit
		}

		return ident, nil
	} else {
		return nil, p.expectedErr(Token_Ident)
	}
}

// parseGenericValue tries to parse first with parseValue, if fails, tries with list,
// if fails, tries with map, if fails, return the error
func (p *Parser) parseGenericValue() (ValueStmt, error) {
	value, err := p.parseValue()
	if err == nil {
		return value, nil
	}

	value, err = p.parseList()
	if err == nil {
		return value, nil
	}

	p.undo() // undo because map expects start with [
	value, err = p.parseMap()
	if err == nil {
		return value, nil
	}

	return nil, err
}

// pareMap parses an expression in the form: [(entry1),(entry2)]
// where "entry" means (key:value). This is a special case of list
func (p *Parser) parseMap() (*MapValueStmt, error) {
	err := p.requireNext(Token_Lbrack)
	if err != nil {
		return nil, err
	}

	m := new(MapValueStmt)
	value, err := p.parseMapEntry()
	if err != nil {
		return nil, err
	}
	m.add(value)

	p.next()
	for p.tok == Token_Comma {
		p.next()
		value, err = p.parseMapEntry()
		if err != nil {
			return nil, err
		}
		m.add(value)
		p.next()
	}

	err = p.require(Token_Rbrack)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// parseMapEntry parses an expression in the form (key:value)
func (p *Parser) parseMapEntry() (*MapEntryStmt, error) {
	// (
	err := p.requireNext(Token_Lparen)
	if err != nil {
		return nil, err
	}

	// parse key
	key, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	// parse :
	p.next()
	err = p.requireNext(Token_Colon)
	if err != nil {
		return nil, err
	}

	// parse value
	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	// )
	err = p.requireMove(Token_Rparen)
	if err != nil {
		return nil, err
	}

	return &MapEntryStmt{
		key:   key,
		value: value,
	}, nil
}

// parseList parses an expression in the form: [value1, value2, value3]
func (p *Parser) parseList() (*ListValueStmt, error) {
	err := p.requireNext(Token_Lbrack)
	if err != nil {
		return nil, err
	}

	list := new(ListValueStmt)
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
			p.undo()

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
		return p.next() // skip the comment
	}

	return nil
}

// nextIs advances one token and returns true if its equal to tok
func (p *Parser) nextIs(tok Token, advance ...bool) bool {
	advanceB := true
	if len(advance) == 1 {
		advanceB = advance[0]
	}

	p.next()
	if advanceB {
		defer p.next()
	}
	return p.tok == tok
}

// undo unscans a token
func (p *Parser) undo(n ...int) {
	length := 2
	if len(n) == 1 {
		length = n[0] + 1
	}

	// unscan 2 because if we unscan 1 when we "next", it will stay in the same token
	p.tokenizer.unscan(length)
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
	return fmt.Errorf("%s -> expected %s, given %q", p.pos.String(), txt, p.lit)
}

func (p *Parser) stringToValue() (primitive Primitive, value interface{}) {
	lit := p.lit

	if lit == nullKeyword {
		return Primitive_Null, nil
	}

	// try with bool
	b, err := strconv.ParseBool(lit)
	if err == nil {
		return Primitive_Bool, b
	}

	return Primitive_Illegal, nil
}
