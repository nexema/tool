package internal

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Parser struct {
	s   *Scanner
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
func (p *Parser) scan(readSpace ...bool) (tok Token, lit string) {
	readSpaceBool := false
	if len(readSpace) > 0 {
		readSpaceBool = readSpace[0]
	}

	// if we have a token on the buffer, return it
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// scan next token
	pos, tok, lit := p.s.Scan(readSpaceBool)

	// save to buffer
	p.buf.pos, p.buf.tok, p.buf.lit = pos, tok, lit
	return
}

func (p *Parser) peek() (tok Token, lit string) {

	// scan next token
	_, tok, lit = p.s.Peek(false)

	return
}

// unscan pushes the previously read token back onto the buffer
func (p *Parser) unscan() {
	p.buf.n = 1
}

// Parse parses the given reader and creates an abstract syntax tree of the input
func (p *Parser) Parse() (*Ast, error) {
	ast := &Ast{
		imports: new(importsStmt),
		types:   new(typesStmt),
	}

	for {
		tok, lit := p.scan()

		// scan until end of stream
		if tok == Token_EOF {
			break
		} else if tok == Token_Backslash {
			p.unscan()
			p.parseComment()
			continue
		}

		// to start, only import or type can be specified
		keyword := inverseKeywordMapping[lit]
		p.unscan()
		if keyword == Keyword_Import {
			importStmt, err := p.parseImport()
			if err != nil {
				return nil, err
			}

			ast.imports.add(importStmt)
		} else if keyword == Keyword_Type || tok == Token_At {
			typeStmt, err := p.parseType()
			if err != nil {
				return nil, err
			}

			ast.types.add(typeStmt)
		} else {
			return nil, p.expectedRawError(`"type" or "import" keywords`, lit)
		}
	}

	return ast, nil
}

// parseImport parses an import keyword
func (p *Parser) parseImport() (*importStmt, error) {

	// read import keyword
	tok, lit := p.scan()
	keyword := inverseKeywordMapping[lit]
	if tok != Token_Keyword || keyword != Keyword_Import {
		return nil, p.expectedKeywordErr(Keyword_Import, lit)
	}

	// read string
	tok, lit = p.scan()
	if tok != Token_String {
		return nil, p.expectedRawError("import path", lit)
	}

	// remove " from string
	lit = lit[1 : len(lit)-1]

	stmt := &importStmt{src: lit}

	// maybe read alias
	tok, lit = p.scan()
	if tok == Token_Keyword && isExactKeyword(lit, Keyword_As) {
		// read alias
		tok, lit = p.scan()
		if tok != Token_Ident {
			return nil, p.expectedRawError("import alias", lit)
		}

		stmt.alias = &lit
	} else {
		p.unscan()
	}

	return stmt, nil
}

// parseType parses a type
func (p *Parser) parseType() (*typeStmt, error) {
	tok, lit := p.scan()

	typeStmt := &typeStmt{
		typeModifier: TypeModifier_Struct,
		fields:       new(fieldsStmt),
	}

	// if tok is @, read metadata for incoming type
	if tok == Token_At {
		mapStmt, err := p.parseMap()
		if err != nil {
			return nil, err
		}

		typeStmt.metadata = mapStmt
		_, lit = p.scan()
	}

	// if the next keyword is not "type", error
	if !isExactKeyword(lit, Keyword_Type) {
		return nil, p.expectedKeywordErr(Keyword_Type, lit)
	}

	// scan type's name
	tok, lit = p.scan()
	if tok != Token_Ident {
		return nil, p.expectedRawError("identifier", lit)
	}
	typeStmt.name = lit

	// cannot create a  struct which name is a keyword
	if isKeyword(lit) {
		return nil, p.keywordGivenErr(lit)
	}

	// read modifier
	tok, lit = p.scan()
	if tok == Token_Keyword {
		modifier, ok := parseTypeModifier(lit)
		if !ok {
			return nil, p.raw(fmt.Sprintf("unknown type modifier: %q", lit))
		}

		typeStmt.typeModifier = modifier
	} else {
		// unscan because we will infer its "{" and the type modifier become "struct" implicitly
		// if its not a "{", the next read should return an error
		p.unscan()
	}

	// read open curly braces "{"
	tok, lit = p.scan()
	if tok != Token_OpenCurlyBraces {
		return nil, p.expectedError(Token_OpenCurlyBraces, lit)
	}

	// from here, start reading fields
	for {
		// read next token
		tok, _ = p.scan()

		if tok == Token_EOF {
			continue
			// return nil, p.expectedError(Token_CloseCurlyBraces, lit)
		}

		// stop reading struct and fields
		if tok == Token_CloseCurlyBraces {
			break
		}

		p.unscan()
		fieldStmt, err := p.parseField(typeStmt.typeModifier)
		if err != nil {
			return nil, err
		}

		typeStmt.fields.add(fieldStmt)
	}

	return typeStmt, nil
}

func (p *Parser) parseField(forModifier TypeModifier) (*fieldStmt, error) {
	field := &fieldStmt{index: -1}

	var tok Token
	var lit string
	for {
		tok, lit = p.scan()

		// expect ident, because of field's index or name
		if tok != Token_Ident {
			return nil, p.expectedError(Token_Ident, lit)
		}

		// if lit can be parsed into an int, its the field's index
		fieldIndex, err := strconv.Atoi(lit)
		if err == nil && field.index == -1 {
			field.index = fieldIndex
		} else {
			// it corresponds to the field's name
			field.name = lit
			break
		}
	}

	if field.index == -1 {
		field.index = 0
	}

	if forModifier == TypeModifier_Enum {
		return field, nil
	}

	// read ":"
	tok, lit = p.scan()
	if tok != Token_Colon {
		return nil, p.expectedError(Token_Colon, lit)
	}

	// read field type
	fieldType, err := p.parseFieldType()
	if err != nil {
		return nil, err
	}

	field.valueType = fieldType

	// maybe read "=" for default value or "@" for metadata
	for {
		tok, _ = p.scan()

		// scan default value
		if tok == Token_Equals {
			defaultValue, err := p.parseValue()
			if err != nil {
				return nil, err
			}

			field.defaultValue = defaultValue
		} else if tok == Token_At {
			m, err := p.parseMap()
			if err != nil {
				return nil, err
			}

			field.metadata = m
		} else {
			p.unscan()
			break
		}
	}

	return field, nil
}

func (p *Parser) parseFieldType() (*valueTypeStmt, error) {
	fieldType := new(valueTypeStmt)

	// read primitive
	tok, lit := p.scan()
	if tok != Token_Keyword && tok != Token_Ident {
		return nil, p.expectedRawError("field type", lit)
	}

	// parse primitive
	primitive, ok := primitiveMapping[lit]
	if !ok {
		// if primitive is not recognized, its because its a custom type
		primitive = Primitive_Type
		fieldType.customTypeName = &lit
	}
	fieldType.primitive = primitive

	// primitive is list or map, expect type arguments
	if fieldType.primitive == Primitive_List || fieldType.primitive == Primitive_Map {
		fieldType.typeArguments = new([]*valueTypeStmt)

		// read (
		tok, lit = p.scan()
		if tok != Token_OpenParens {
			if fieldType.primitive == Primitive_List {
				return nil, p.raw(fmt.Sprintf("lists expect one type argument, given: %s", lit))
			} else {
				if fieldType.primitive == Primitive_Map {
					return nil, p.raw(fmt.Sprintf("maps expect two type arguments, given: %s", lit))
				}
			}
		}

		firstArgument, err := p.parseFieldType()
		if err != nil {
			return nil, err
		}
		(*fieldType.typeArguments) = append((*fieldType.typeArguments), firstArgument)

		// if map, read next
		if fieldType.primitive == Primitive_Map {
			tok, lit = p.scan()
			if tok != Token_Comma {
				return nil, p.expectedError(Token_Comma, lit)
			}

			secondArgument, err := p.parseFieldType()
			if err != nil {
				return nil, err
			}

			(*fieldType.typeArguments) = append((*fieldType.typeArguments), secondArgument)
		}

		// read close parens )
		tok, lit = p.scan()
		if tok != Token_CloseParens {
			return nil, p.expectedError(Token_CloseParens, lit)
		}
	}

	// maybe read "?" for nullable
	tok, _ = p.scan()
	if tok != Token_QuestionMark {
		p.unscan()
	} else {
		fieldType.nullable = true
	}

	return fieldType, nil
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
			identifier, err := p.parseValue()
			if err != nil {
				return nil, err
			}

			primitive := identifier.Primitive()
			if primitive == Primitive_Map || primitive == Primitive_List {
				return nil, p.raw("lists cannot contain nested lists or maps")
			}

			// add identifier to stmt
			l.add(identifier.(*identifierStmt))
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
					identifier, err := p.parseValue()
					if err != nil {
						return nil, err
					}

					primitive := identifier.Primitive()
					if primitive == Primitive_Map || primitive == Primitive_List {
						return nil, p.raw("lists cannot contain nested lists or maps")
					}

					if entry.key == nil {
						entry.key = identifier.(*identifierStmt)

						// next must be colon
						tok, _ = p.scan()
						if tok != Token_Colon {
							return nil, p.raw("key-value pair must be in the format key:value")
						}
					} else if entry.value == nil {
						entry.value = identifier.(*identifierStmt)
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

func (p *Parser) parseValue() (baseIdentifierStmt, error) {
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
			valueType: &valueTypeStmt{primitive: Primitive_String},
		}, nil
	} else if tok == Token_Ident {
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
					// if not a bool, try with null
					primitive = primitiveMapping[lit]
					if primitive == Primitive_Null {
						value = nil
					} else {
						// try to parse a custom enum value
						p.unscan()
						stmt, err := p.parseCustomTypeValue()
						if err != nil {
							return nil, p.raw(fmt.Sprintf("unknown primitive %s", lit))
						}

						return stmt, nil
					}
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
			value: value,
			valueType: &valueTypeStmt{
				primitive: primitive,
			},
		}, nil
	} else if tok == Token_OpenBrackets { // for list or map
		//p.unscanBuf()
		p.unscan()

		// if next token is (, then is a map
		tok, _ = p.peek()
		if tok == Token_OpenParens {
			m, err := p.parseMap()
			if err != nil {
				return nil, err
			}

			return m, nil
		} else {

			l, err := p.parseList()
			if err != nil {
				return nil, err
			}

			return l, nil
		}
	} else {
		return nil, p.expectedRawError("primitive", lit)
	}
}

func (p *Parser) parseCustomTypeValue() (*customTypeIdentifierStmt, error) {
	tok, lit := p.scan()
	if tok != Token_Ident {
		return nil, p.expectedError(Token_Ident, lit)
	}

	// custom type values only allows enums, so its format becomes alias.enumName.value
	// alias is optional, so for now we only split lit
	toks := strings.Split(lit, ".")
	if len(toks) < 2 || len(toks) > 3 { // at least two, enumName and value and no more than 3: alias, name and value
		return nil, p.expectedRawError("enum name and value", lit)
	}

	var enumIdentifier, value string

	if len(toks) == 3 {
		enumIdentifier = toks[0] + "." + toks[1]
		value = toks[2]
	} else {
		enumIdentifier = toks[0]
		value = toks[1]
	}

	return &customTypeIdentifierStmt{
		customTypeName: enumIdentifier,
		value:          value,
	}, nil
}

func (p *Parser) parseComment() (*commentStmt, error) {
	backlashCount := 0
	multine := false
	commentStarted := false
	commentType := singleline

	sb := new(strings.Builder)
	for {
		tok, lit := p.scan(true)
		if tok == Token_EOF {
			if !commentStarted {
				return nil, p.raw("comments must start with double // or with /* for multine comments")
			} else if commentStarted && multine {
				return nil, p.raw("multine comments must be closed with */")
			} else {
				return &commentStmt{
					text:        sb.String(),
					commentType: commentType,
				}, nil
			}
		} else if tok == Token_Backslash {
			// if last token was a / too, add to backlash count
			// in order to determine the type of comment
			if !commentStarted {
				backlashCount++

				continue
			}
		} else if tok == Token_Asterisk {
			tok, lit = p.scan() // if next tok is /, its closing the comment
			if tok == Token_Backslash {
				if multine {
					return &commentStmt{
						text:        sb.String(),
						commentType: commentType,
					}, nil
				}
			} else {
				p.unscan()
				if backlashCount > 0 {
					// start the comment
					multine = true
					commentStarted = true
					commentType = multiline
					continue
				}
			}
		} else if tok == Token_Newline {
			if multine {
				sb.WriteString(lit)
				continue
			} else {
				return &commentStmt{
					text:        sb.String(),
					commentType: commentType,
				}, nil
			}
		}

		sb.WriteString(lit)

		if !commentStarted && backlashCount > 1 {
			commentStarted = true
		}
	}
}
