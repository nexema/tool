package internal

import (
	"bytes"
	"fmt"
	"strconv"
)

// Parser takes tokens from a Tokenizer and produces a Ast.
// Parser maintains a list of read tokens, called cache, with their literal and position, in order to allow
// unscanning. When .consume() is called, the newly token is pushed onto the cache and its position is reset to the last
// added element. Cache provides a method to set the read position, causing future calls to Cache.Next() to return
// cached tokens from that position instead of read from the underlying Tokenizer.
type Parser struct {
	tokenizer *Tokenizer

	buf   parserBuf         // the current buffer
	cache *Cache[parserBuf] // the cache of read tokens

	comments *map[int]*CommentStmt // map of any comment found. Map's key is the line start
	imports  *[]*ImportStmt
	types    *[]*TypeStmt
}

type parserBuf struct {
	tok Token    // current token
	pos Position // current token's position
	lit string   // current token's literal
}

// NewParser builds a new Parser
func NewParser(buf *bytes.Buffer) *Parser {
	tokenizer := NewTokenizer(buf)
	return &Parser{
		tokenizer: tokenizer,
		buf: parserBuf{
			tok: Token_Illegal,
			pos: tokenizer.pos,
		},
		cache:    NewCache[parserBuf](),
		comments: &map[int]*CommentStmt{},
		imports:  new([]*ImportStmt),
		types:    new([]*TypeStmt),
	}
}

// Parse parses the given input when creating the Parser and returns
// the corresponding Ast
func (p *Parser) Parse() (*Ast, error) {
	p.next() // read initial token

	// scan any import
	if p.buf.tok == Token_Import && p.nextIs(Token_Colon, false) {
		// comeback one pos to get into :
		p.undo(len(importKeyword))
		err := p.parseImportGroup()
		if err != nil {
			return nil, err
		}
	}

	// scan types
	for p.buf.tok == Token_At || p.buf.tok == Token_Type {
		typeStmt, err := p.parseType()
		if err != nil {
			return nil, err
		}

		(*p.types) = append((*p.types), typeStmt)
		p.next()
	}

	return &Ast{
		Imports: p.imports,
		Types:   p.types,
	}, nil
}

// parseImportGroup parses a expression in the form import:\n "import1"\n "import2"
func (p *Parser) parseImportGroup() error {
	err := p.requireSeq(Token_Import, Token_Colon)
	if err != nil {
		return err
	}

	for p.buf.tok == Token_String {
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
		Path: &IdentifierStmt{Lit: p.buf.lit},
	}
	p.next()

	// may parse AS alias
	if p.buf.tok == Token_As {
		p.next()
		alias, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		stmt.Alias = alias
	}

	return stmt, nil
}

// parseField parses a FieldStmt
func (p *Parser) parseField() (*FieldStmt, error) {
	stmt := new(FieldStmt)

	// if until this moment any comment has been read, add as documentation
	if len(*p.comments) > 0 {
		stmt.Documentation = p.getComments()
	}

	// maybe read field index
	if p.buf.tok == Token_Int {
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		stmt.Index = value
		p.next()
	}

	// field's name
	if p.buf.tok == Token_Ident {
		value, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		stmt.Name = value
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
	if p.buf.tok == Token_Ident {
		value, err := p.parseValueTypeStmt()
		if err != nil {
			return nil, err
		}

		stmt.ValueType = value
	} else {
		return nil, p.expectedErr(Token_Ident)
	}

	if p.buf.tok == Token_Assign { // default value
		p.next()
		value, err := p.parseGenericValue()
		if err != nil {
			return nil, err
		}

		stmt.DefaultValue = value
		p.next()
	}

	if p.buf.tok == Token_At { // metadata
		p.next()
		value, err := p.parseMap()
		if err != nil {
			return nil, err
		}

		stmt.Metadata = value
		p.next()
	}

	return stmt, nil
}

// parseEnumField parses a FieldStmt but using Enum's syntax.
// Enum's fields uses the same signature as Struct or Union's fields but
// type is not allowed nor default value
func (p *Parser) parseEnumField() (*FieldStmt, error) {
	stmt := new(FieldStmt)

	// if until this moment any comment has been read, add as documentation
	if len(*p.comments) > 0 {
		stmt.Documentation = p.getComments()
	}

	// maybe read field index
	if p.buf.tok == Token_Int {
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		stmt.Index = value
		p.next()
	}

	// field's name
	if p.buf.tok == Token_Ident {
		value, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		stmt.Name = value
	} else {
		return nil, p.expectedErr(Token_Ident)
	}

	if p.buf.tok == Token_At { // metadata
		p.next()
		value, err := p.parseMap()
		if err != nil {
			return nil, err
		}

		stmt.Metadata = value
		p.next()
	}

	return stmt, nil
}

// parseType parses a TypeStmt
func (p *Parser) parseType() (*TypeStmt, error) {
	stmt := new(TypeStmt)

	// check if there is any comment that is a possible documentation comment
	stmt.Documentation = p.getValidCommentsFor(p.buf.pos.line)

	if p.buf.tok == Token_At { // metadata present
		p.next()
		// parse map
		mapStmt, err := p.parseMap()
		if err != nil {
			return nil, err
		}

		stmt.Metadata = mapStmt
		p.next()
	}

	if p.buf.tok != Token_Type {
		return nil, p.expectedErr(Token_Type)
	}
	p.next()

	// read name
	ident, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	stmt.Name = ident
	// p.next()

	// if read {, then modifier is struct
	if p.buf.tok == Token_Lbrace {
		stmt.Modifier = Token_Struct
	} else {
		// read modifier
		if !p.buf.tok.IsModifier() {
			return nil, p.expectedGiven("expected type modifier")
		}

		stmt.Modifier = p.buf.tok

		p.next()
		err := p.require(Token_Lbrace)
		if err != nil {
			return nil, err
		}
	}
	// read fields until "}"
	p.next()
	for {
		if p.buf.tok == Token_Rbrace {
			goto exit
		}

		if p.buf.tok == Token_EOF {
			break
		}

		if stmt.Fields == nil {
			stmt.Fields = new([]*FieldStmt)
		}

		var fieldStmt *FieldStmt
		var err error
		if stmt.Modifier == Token_Enum {
			fieldStmt, err = p.parseEnumField()
		} else {
			fieldStmt, err = p.parseField()
		}

		if err != nil {
			return nil, err
		}

		(*stmt.Fields) = append((*stmt.Fields), fieldStmt)
		// p.undo()
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
	stmt.Ident, err = p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	// if we read parens, it contains type arguments
	if p.buf.tok == Token_Lparen {
		p.next()
		stmt.TypeArguments = new([]*ValueTypeStmt)

		// read first one
		valueType, err := p.parseValueTypeStmt()
		if err != nil {
			return nil, err
		}

		(*stmt.TypeArguments) = append((*stmt.TypeArguments), valueType)

		for p.buf.tok == Token_Comma {
			p.next()
			valueType, err := p.parseValueTypeStmt()
			if err != nil {
				return nil, err
			}

			(*stmt.TypeArguments) = append((*stmt.TypeArguments), valueType)
		}

		// must read )
		err = p.require(Token_Rparen)
		if err != nil {
			return nil, err
		}

		p.next()
	}

	// nullable
	if p.buf.tok == Token_Nullable {
		stmt.Nullable = true
		p.next()
	}

	return stmt, err
}

// parseIdentifier parses an IdentifierStmt. It can be string,
// bool, or any other primitive, enum's values, etc.
func (p *Parser) parseIdentifier() (*IdentifierStmt, error) {
	if p.buf.tok == Token_Ident {
		ident := new(IdentifierStmt)
		lit := p.buf.lit
		p.next()

		// while found a period, it can be import alias, type's name or type's value
		if p.buf.tok == Token_Period {
			p.next()
			err := p.require(Token_Ident)
			if err != nil {
				return nil, err
			}

			ident.Alias = lit
			ident.Lit = p.buf.lit
			p.next()
		} else {
			ident.Lit = lit
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
	for p.buf.tok == Token_Comma {
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
		Key:   key,
		Value: value,
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
	for p.buf.tok == Token_Comma {
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

// parseValue parses any primitive raw value
func (p *Parser) parseValue() (ValueStmt, error) {
	if p.buf.tok == Token_String {
		return &PrimitiveValueStmt{Primitive: Primitive_String, RawValue: p.buf.lit}, nil
	} else if p.buf.tok == Token_Ident {
		kind, value := p.stringToValue()
		if kind != Primitive_Illegal {
			return &PrimitiveValueStmt{Primitive: kind, RawValue: value}, nil
		}

		// if its illegal, try with enum, it needs the name, and a value.
		enumName, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		// if the next token is a period, we found a declaration in the form
		// alias.EnumName.enum_value, otherwise, enumName contains the EnumName.enum_value
		if p.buf.tok == Token_Period {
			p.next()
			enumValue, err := p.parseIdentifier()
			if err != nil {
				return nil, err
			}

			return &TypeValueStmt{
				TypeName: enumName,
				RawValue: enumValue,
			}, nil
		} else {
			p.undo() // undo the invalid read token

			// alias becomes the name, and lit the value
			return &TypeValueStmt{
				TypeName: &IdentifierStmt{Lit: enumName.Alias},
				RawValue: &IdentifierStmt{Lit: enumName.Lit},
			}, nil
		}

	} else if p.buf.tok == Token_Float {
		f, _ := strconv.ParseFloat(p.buf.lit, 64)
		return &PrimitiveValueStmt{RawValue: f, Primitive: Primitive_Float64}, nil
	} else if p.buf.tok == Token_Int {
		i, _ := strconv.ParseInt(p.buf.lit, 10, 64)
		return &PrimitiveValueStmt{RawValue: i, Primitive: Primitive_Int64}, nil
	} else {
		return nil, p.expectedGiven("string, int, float or identifier")
	}
}

// next sets the current token to the newly read from consume, if stack is empty.
// Otherwise, it will pop the stack until zero elements
// Any comment that is encountered, is saved for later use
func (p *Parser) next() error {
	var err error
	if p.cache.NextHas() {
		buf := p.cache.Advance()
		p.buf = *buf
	} else {
		err = p.consume()
	}

	if err != nil {
		return err
	}

	// save any comment
	if p.buf.tok == Token_Comment {
		comment, err := p.parseComment()
		if err != nil {
			return err
		}

		// if the comment is on the same line that another token, its not documentation
		before := p.cache.Before()
		if before == nil || before.pos.line != comment.lineStart {
			(*p.comments)[comment.lineStart] = comment
		}

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
	return p.buf.tok == tok
}

// undo unscans a token
func (p *Parser) undo(n ...int) {
	count := 1
	if len(n) == 1 {
		count = n[0]
	}

	// do not unscan from tokenizer, instead, pop the stack
	// p.tokenizer.unscan(length)
	// p.next()

	p.cache.Back(count)
	p.buf = *p.cache.Current()
}

// consume reads a token from the underlying tokenizer and saves it in the Parser
func (p *Parser) consume() error {
	var err error
	p.buf.pos, p.buf.tok, p.buf.lit, err = p.tokenizer.Scan()
	p.cache.Push(p.buf)
	return err
}

// parseComment parses the next comment
func (p *Parser) parseComment() (*CommentStmt, error) {
	// scan for /*-style comments to report line end
	lineEnd := p.buf.pos.line
	if p.buf.lit[1] == '*' {
		for _, ch := range p.buf.lit {
			if ch == '\n' {
				lineEnd++
			}
		}
	}

	return &CommentStmt{
		Text:      p.buf.lit,
		posStart:  p.buf.pos.offset,
		posEnd:    p.buf.pos.offset + len(p.buf.lit),
		lineStart: p.buf.pos.line,
		lineEnd:   lineEnd,
	}, nil
}

// require returns an error if the next token does not match tok
func (p *Parser) require(tok Token) error {

	// try to move one position
	if p.buf.tok == Token_Illegal {
		p.next()
	}

	if p.buf.tok != tok {
		return p.expectedErr(tok)
	}
	return nil
}

// requireNext does the same as require but moves to the next token after matching
func (p *Parser) requireNext(tok Token) error {

	// try to move one position
	if p.buf.tok == Token_Illegal {
		p.next()
	}

	if p.buf.tok != tok {
		return p.expectedErr(tok)
	}

	p.next()
	return nil
}

// requireMove does the same as require but tries to move one position trying to fetch the required token
func (p *Parser) requireMove(tok Token) error {

	// try to move one position
	if p.buf.tok == Token_Illegal {
		p.next()
	}

	if p.buf.tok != tok {
		p.next()

		if p.buf.tok != tok {
			return p.expectedErr(tok)
		}
	}
	return nil
}

// requireSeq checks if the next sequence of two tokens matches first and second
func (p *Parser) requireSeq(first Token, second Token) error {
	// try to move to a legal position
	if p.buf.tok == Token_Illegal {
		p.next()
	}

	if p.buf.tok != first {
		return p.expectedErr(first)
	}

	p.next()
	if p.buf.tok != second {
		return p.expectedErr(second)
	}

	p.next()
	return nil
}

func (p *Parser) expectedErr(tok Token) error {
	return fmt.Errorf("%s -> expected %q, given %q (%s)", p.buf.pos.String(), tok.String(), p.buf.tok.String(), p.buf.lit)
}

func (p *Parser) expectedGiven(txt string) error {
	return fmt.Errorf("%s -> expected %s, given %q", p.buf.pos.String(), txt, p.buf.lit)
}

func (p *Parser) stringToValue() (primitive Primitive, value interface{}) {
	lit := p.buf.lit

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

func (p *Parser) getValidCommentsFor(line int) *[]*CommentStmt {
	/* the procedure is: check if there is any comment at line-1.
	if any comment, check at (line-1)-1, and so on until no more comments are encountered
	*/

	comments := new([]*CommentStmt)
	currLine := line - 1
	if len(*p.comments) > 0 {
		for {
			comment, ok := (*p.comments)[currLine]
			if !ok {
				break
			}

			if comment.lineStart == currLine {
				(*comments) = append(*comments, comment)
				currLine--
			} else {
				break
			}
		}
	}

	// reverse the list
	for i, j := 0, len(*comments)-1; i < j; i, j = i+1, j-1 {
		(*comments)[i], (*comments)[j] = (*comments)[j], (*comments)[i]
	}

	// new map
	p.comments = &map[int]*CommentStmt{}

	return comments
}

// getComments returns p.comments as an array then clears the map
func (p *Parser) getComments() *[]*CommentStmt {
	list := make([]*CommentStmt, len(*p.comments))

	idx := 0
	for _, elem := range *p.comments {
		list[idx] = elem
		idx++
	}

	// new map
	p.comments = &map[int]*CommentStmt{}

	return &list
}
