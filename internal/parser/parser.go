package parser

import (
	"io"
	"strconv"

	"github.com/tidwall/btree"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
	"tomasweigenast.com/nexema/tool/internal/tokenizer"
)

type Parser struct {
	tokenizer             tokenizer.Tokenizer
	file                  *File
	currentToken          *tokenBuf
	nextToken             *tokenBuf
	errors                *ParserErrorCollection
	eof                   bool
	annotationsOrComments *btree.Map[int, *[]annotationOrComment]
}

type tokenBuf struct {
	token    *token.Token
	position *reference.Pos
}

func NewParser(input io.Reader, file *File) *Parser {
	return &Parser{
		file:         file,
		eof:          false,
		currentToken: nil,
		nextToken:    nil,
		tokenizer:    *tokenizer.NewTokenizer(input),
	}
}

func (self *Parser) Errors() *ParserErrorCollection {
	return self.errors
}

func (self *Parser) Parse() *Ast {
	// read "use" statements
	var useStmts []UseStmt
	for self.currentTokenIs(token.Use) && self.nextTokenIs(token.String) {
		stmt := self.parseUseStmt()
		if stmt == nil {
			break
		}

		useStmts = append(useStmts, *stmt)
		self.next()
	}

	// read typeStmts
	var typeStmts []TypeStmt
	for self.currentTokenIs(token.Type) {
		self.next()
		stmt := self.parseTypeStmt()
		if stmt == nil {
			break
		}

		typeStmts = append(typeStmts, *stmt)
		self.next()
	}

	if self.currentToken != nil {
		self.pushError(ErrUnexpectedToken{Got: *self.currentToken.token}, self.currentToken.position)
	}

	return &Ast{
		File:           self.file,
		UseStatements:  useStmts,
		TypeStatements: typeStmts,
	}
}

// func (self *Parser) Begin() {
// 	self.Reset()
// 	self.next()
// }

// Reset initializes the parser reading one token into the buffer
func (self *Parser) Reset() *ParserError {
	result := self.consume()
	if result != nil {
		return NewParserErr(ErrTokenizer{*result}, self.getReference())
	}

	// advance one more to get nextToken into currentToken
	result = self.consume()
	if result != nil {
		return NewParserErr(ErrTokenizer{*result}, self.getReference())
	}

	self.eof = false
	self.resetAnnotationCommentsMap()
	self.errors = newParserErrorCollection()

	return nil
}

// parseTypeStmt parses a type statement.
func (self *Parser) parseTypeStmt() *TypeStmt {
	// "type" keyword already read

	// any comment or annotation read until here is added to the type
	var annotations []AnnotationStmt = nil
	var comments []CommentStmt = nil
	if self.currentToken != nil {
		// any comment or annotation read until here, while they appear one line before each other, must be added as doc
		currentLine := self.currentToken.position.Line
		arr := self.getAnnotationsAndComments(currentLine)
		unwrapAnnotationsOrComments(arr, &annotations, &comments)
	}

	// read type name
	typeName := self.parseIdent()
	if typeName == nil {
		return nil
	}

	self.next()

	// read type modifier or "extends" keyword
	var baseType *DeclStmt
	modifier := token.Struct
	currentToken := self.currentToken

	if currentToken == nil {
		return nil
	}

	switch currentToken.token.Kind {
	case token.Extends:
		self.next()
		baseType = self.parseDeclStmt(true)
		if baseType == nil {
			self.pushError(ErrExpectedIdentifier{*currentToken.token})
			return nil
		}

	case token.Struct, token.Enum, token.Union, token.Base:
		modifier = currentToken.token.Kind

	default:
		return nil
	}

	// now require {
	if !self.expectToken(token.Lbrace) {
		return nil
	}

	// read fields until } or "defaults" keyword
	var fields []FieldStmt
	var defaults []AssignStmt = nil

	for !self.currentTokenIs(token.Rbrace) {
		if self.currentToken == nil {
			break
		}

		switch self.currentToken.token.Kind {
		case token.Defaults:
			self.next()
			defaults = self.parseDefaultsBlock()

			// after reading the defaults block, nothing can be declared, so expect closing
			if !self.expectToken(token.Rbrace) {
				return nil
			}

		case token.Rbrace:
			break // exit

		default:
			// read field
			self.next()
			fieldStmt := self.parseFieldStmt(modifier == token.Enum)
			if fieldStmt == nil {
				break
			}

			fields = append(fields, *fieldStmt)
		}
	}

	if !self.currentTokenIs(token.Rbrace) {
		return nil
	}

	return &TypeStmt{
		BaseType:      baseType,
		Name:          *typeName,
		Modifier:      modifier,
		Documentation: comments,
		Annotations:   annotations,
		Fields:        fields,
		Defaults:      defaults,
	}
}

// parseDefaultsBlock parses a block of declarations.
func (self *Parser) parseDefaultsBlock() []AssignStmt {
	// "defaults" keyword already read
	if !self.expectCurrentToken(token.Lbrace) {
		return nil
	}

	assigments := make([]AssignStmt, 0)
	for !self.currentTokenIs(token.Rbrace) {
		stmt := self.parseAssignStmt()
		if stmt == nil {
			return nil
		}

		assigments = append(assigments, *stmt)
		self.next()
	}

	return assigments
}

// parseFieldStmt parses a declaration in the form:
//
// (index)     [ident]     [decl]
// field_index field_name  value_type.
//
// Any comment or annotation that was read until this method call, will be added as
// documentation (if comment line is self.tokenizer.currentLine-1) and annotations, respectively.
func (self *Parser) parseFieldStmt(isEnum bool) *FieldStmt {

	var annotations []AnnotationStmt = nil
	var comments []CommentStmt = nil
	if self.currentToken != nil {
		// any comment or annotation read until here, while they appear one line before each other, must be added as doc
		currentLine := self.currentToken.position.Line
		arr := self.getAnnotationsAndComments(currentLine)
		unwrapAnnotationsOrComments(arr, &annotations, &comments)
	}

	if self.currentToken == nil {
		self.pushError(ErrUnexpectedEOF{})
		return nil
	}

	var fieldIndex *IdentStmt

	// if current token is a number, its the index
	if self.currentTokenIs(token.Integer) {
		fieldIndex = &IdentStmt{
			Token: *self.currentToken.token,
			Pos:   *self.currentToken.position,
		}
		self.next()
	}

	// read field name
	fieldName := self.parseIdent()
	if fieldName == nil {
		return nil
	}

	var fieldType *DeclStmt
	if !isEnum {
		// read type declaration
		self.next()
		fieldType = self.parseDeclStmt(false)
		if fieldType == nil {
			return nil
		}
	}

	return &FieldStmt{
		Index:         fieldIndex,
		Name:          *fieldName,
		ValueType:     fieldType,
		Documentation: comments,
		Annotations:   annotations,
	}
}

// parseUseStmt parses a statement in the following form:
//
// use "path/to/my/package"
// use "path/to/my/package" as my_pkg
func (self *Parser) parseUseStmt() *UseStmt {
	self.next()

	// now we need the path as a string
	if self.currentToken == nil {
		self.pushError(ErrUnexpectedValue{"literal with import path", *token.Token_EOF})
		return nil
	}

	if self.currentTokenIs(token.String) {
		literal := self.parseLiteral()
		if literal == nil {
			return nil
		} else {

			useStmt := &UseStmt{
				Token: *token.NewToken(token.Use),
				Path:  *literal,
				Alias: nil,
			}

			// maybe alias
			if self.nextTokenIs(token.As) {
				self.next()

				if self.nextTokenIs(token.Ident) {
					self.next()
					ident := self.parseIdent()
					useStmt.Alias = ident
				}
			}

			return useStmt
		}
	}

	return nil
}

// parseDeclStmt parses a statement in the following form:
//
// string
// int
// list(string)
// map(int, string)
// MyType
func (self *Parser) parseDeclStmt(reportErr bool) *DeclStmt {
	if self.currentToken == nil {
		if reportErr {
			self.pushError(ErrExpectedDeclaration{*token.Token_EOF})
		}
		return nil
	}

	currentToken := *self.currentToken.token
	currentPos := *self.currentToken.position

	if currentToken.Kind == token.Ident {
		if self.nextToken == nil {
			return &DeclStmt{
				Token:    currentToken,
				Pos:      currentPos,
				Args:     nil,
				Alias:    nil,
				Nullable: false,
			}
		} else {
			nextToken := *self.nextToken
			switch nextToken.token.Kind {
			// if next token is (, it may contain type arguments
			case token.Lparen:
				self.next() // advance to get self.currentToken set to lparen

				args := make([]DeclStmt, 0)

				// while we dont reach ), read DeclStmts, separated by commas
				for !self.nextTokenIs(token.Rparen) {
					self.next()
					decl := self.parseDeclStmt(true)
					if decl == nil {
						break
					}

					args = append(args, *decl)

					self.next()

					if self.currentTokenIs(token.Comma) {
						continue
					} else if self.currentTokenIs(token.Rparen) {
						break
					} else {
						self.reportExpectedNextTokenErr(token.Rparen)
						break
					}
				}

				if self.currentToken == nil {
					self.reportExpectedCurrentTokenErr(token.Rparen)
					return nil
				}

				endPos := *self.currentToken.position
				return &DeclStmt{
					Token:    currentToken,
					Pos:      *reference.NewPos(currentPos.Start, endPos.End, currentPos.Line, endPos.Endline),
					Args:     args,
					Alias:    nil,
					Nullable: self.nextTokenIsMove(token.QuestionMark),
				}

			// if its a period, maybe it contains an alias declaration
			case token.Period:
				// move twice to advance dot and set current token to next identifier
				self.next()
				self.next()

				ident := self.parseIdent()
				if ident == nil {
					return nil
				}

				return &DeclStmt{
					Token: ident.Token,
					Pos:   *reference.NewPos(currentPos.Start, ident.Pos.End, currentPos.Line, ident.Pos.Endline),
					Args:  nil,
					Alias: &IdentStmt{
						Token: currentToken,
						Pos:   currentPos,
					},
					Nullable: self.nextTokenIsMove(token.QuestionMark),
				}

			default:
				return &DeclStmt{
					Token:    currentToken,
					Pos:      *reference.NewPos(currentPos.Start, currentPos.End, currentPos.Line, currentPos.Endline),
					Args:     nil,
					Alias:    nil,
					Nullable: self.nextTokenIsMove(token.QuestionMark),
				}
			}
		}
	} else {
		self.pushError(ErrExpectedIdentifier{currentToken})
		return nil
	}
}

// parseAnnotationStmt parses a declaration in the following form:
//
// #left = right
func (self *Parser) parseAnnotationStmt() *AnnotationStmt {
	assignStmt := self.parseAssignStmt()
	if assignStmt == nil {
		return nil
	}

	pos := assignStmt.Pos
	pos.Start--
	return &AnnotationStmt{
		Token:     *token.NewToken(token.Hash),
		Assigment: *assignStmt,
		Pos:       pos,
	}
}

// parseAssignStmt parses a statement in the following form:
//
// ident = literal
func (self *Parser) parseAssignStmt() *AssignStmt {
	ident := self.parseIdent()
	if ident == nil {
		return nil
	}

	if !self.expectToken(token.Assign) {
		return nil
	}

	self.next()

	literal := self.parseLiteral()
	if literal == nil {
		return nil
	}

	return &AssignStmt{
		Token: *token.NewToken(token.Assign),
		Left:  *ident,
		Right: *literal,
		Pos:   *reference.NewPos(ident.Pos.Start, literal.Pos.End, ident.Pos.Line, literal.Pos.Endline),
	}
}

// parseIdent parses an identifier.
func (self *Parser) parseIdent() *IdentStmt {
	if self.currentToken == nil {
		self.pushError(ErrExpectedIdentifier{*token.Token_EOF})
		return nil
	}

	if self.currentToken.token.Kind == token.Ident {
		return &IdentStmt{
			Token: *self.currentToken.token,
			Pos:   *self.currentToken.position,
		}
	}

	return nil
}

// parseLiteral parses a token literal, like strings, numbers, booleans, lists and maps
func (self *Parser) parseLiteral() *LiteralStmt {
	if self.currentToken == nil {
		self.pushError(ErrExpectedLiteral{*token.Token_EOF})
		return nil
	}

	literalToken := *self.currentToken.token
	tokenPos := *self.currentToken.position

	var literalKind LiteralKind
	literal := literalToken.Literal
	switch literalToken.Kind {
	case token.String:
		literalKind = StringLiteral{literal}

	case token.Integer:
		num, err := strconv.ParseInt(literal, 10, 64)
		if err != nil {
			self.pushError(ErrNumberParse{err, literal})
			return nil
		}

		literalKind = IntLiteral{num}

	case token.Decimal:
		num, err := strconv.ParseFloat(literal, 64)
		if err != nil {
			self.pushError(ErrNumberParse{err, literal})
			return nil
		}

		literalKind = FloatLiteral{num}

	case token.Ident:
		if literal == "true" {
			literalKind = BooleanLiteral{true}
		} else if literal == "false" {
			literalKind = BooleanLiteral{false}
		} else {
			self.pushError(ErrInvalidLiteral{literalToken})
			return nil
		}

	// list literal
	case token.Lbrack:
		self.next()
		literalKind = make(ListLiteral, 0)

		// read literals until ] is found
		for !self.currentTokenIs(token.Rbrack) {
			literal := self.parseLiteral()
			if literal == nil {
				break
			}

			literalKind = append(literalKind.(ListLiteral), *literal)

			// require comma if more tokens to read
			if self.nextTokenIs(token.Comma) {
				self.next()
			} else if self.nextTokenIs(token.Rbrack) {
				self.next()
				continue
			} else {
				self.reportExpectedNextTokenErr(token.Rbrack)
				break
			}

			self.next()
		}

		// expect closing ]
		if !self.currentTokenIs(token.Rbrack) {
			self.reportExpectedCurrentTokenErr(token.Rbrack)
			return nil
		}

		endPos := self.currentToken.position
		tokenPos = *reference.NewPos(tokenPos.Start, endPos.End, tokenPos.Line, endPos.Endline)
		literalToken = *token.NewToken(token.List)

	// map literal
	case token.Lbrace:
		self.next()

		literalKind = make(MapLiteral, 0)
		// read literals until ] is found
		for !self.currentTokenIs(token.Rbrace) {
			// parse key
			keyLiteral := self.parseLiteral()
			if keyLiteral == nil {
				break
			}

			// require colon
			if !self.expectToken(token.Colon) {
				break
			}

			// move to the next token skipping the colon
			self.next()

			// parse value
			valueLiteral := self.parseLiteral()
			if valueLiteral == nil {
				break
			}

			// push key and value
			literalKind = append(literalKind.(MapLiteral), MapEntry{*keyLiteral, *valueLiteral})

			// require comma if more tokens to read
			if self.nextTokenIs(token.Comma) {
				self.next()
			} else if self.nextTokenIs(token.Rbrace) {
				self.next()
				continue
			} else {
				self.reportExpectedNextTokenErr(token.Rbrace)
				break
			}

			self.next()
		}

		// expect closing }
		if !self.currentTokenIs(token.Rbrace) {
			self.reportExpectedCurrentTokenErr(token.Rbrace)
			return nil
		}

		endPos := self.currentToken.position
		tokenPos = *reference.NewPos(tokenPos.Start, endPos.End, tokenPos.Line, endPos.Endline)
		literalToken = *token.NewToken(token.Map)

	default:
		self.pushError(ErrInvalidLiteral{literalToken})
		return nil
	}

	return &LiteralStmt{
		Token: literalToken,
		Kind:  literalKind,
		Pos:   tokenPos,
	}
}

// next reads the next token, skipping comments and saving them for later use
func (self *Parser) next() {
	err := self.consume()
	if err != nil {
		self.pushError(ErrTokenizer{})
		return
	}

	// stop the parsing process
	if self.currentToken == nil {
		self.eof = true
		return
	}

	// read comments until no more are found.
	// here, currentToken will always contain something if .consume() fails, it returns an error
	currToken := self.currentToken
	switch currToken.token.Kind {
	case token.Comment:
		line := currToken.position.Endline
		self.pushComment(line, &CommentStmt{Token: *currToken.token, Pos: *currToken.position})

		self.next()
		return

	case token.CommentMultiline:
		self.next()
		return

	case token.Hash:
		self.next()
		annotationStmt := self.parseAnnotationStmt()
		if annotationStmt != nil {
			line := annotationStmt.Pos.Endline
			self.pushAnnotation(line, annotationStmt)
			self.next()
			return
		}
	}
}

// consume reads token twice from the tokenizer to store current and next.
func (self *Parser) consume() *tokenizer.TokenizerErr {
	self.currentToken = self.nextToken
	tok, pos, err := self.tokenizer.Next()
	if err != nil {
		return err
	}

	if tok.IsEOF() {
		self.nextToken = nil
		return nil
	}

	self.nextToken = &tokenBuf{tok, pos}
	return nil
}

// expectToken ensures the next_token kind is [token].
// If true, it advances the next one, otherwise, return false and reports error.
func (self *Parser) expectToken(token token.TokenKind) bool {
	if self.nextTokenIs(token) {
		self.next()
		return true
	}

	self.reportExpectedNextTokenErr(token)
	return false
}

// expectCurrentToken ensures the currentToken kind is [token].
func (self *Parser) expectCurrentToken(token token.TokenKind) bool {
	if self.currentTokenIs(token) {
		self.next()
		return true
	}

	self.reportExpectedCurrentTokenErr(token)
	return false
}

// nextTokenIs returns true if the next token's kind is [token]
func (self *Parser) nextTokenIs(token token.TokenKind) bool {
	if self.nextToken == nil {
		return false
	}

	return self.nextToken.token.Kind == token
}

// currentTokenIs returns true if the current token is [token]
func (self *Parser) currentTokenIs(token token.TokenKind) bool {
	if self.currentToken == nil {
		return false
	}

	return self.currentToken.token.Kind == token
}

// nextTokenIsMove reurns true if the next token's kind is [token] and advance one if true
func (self *Parser) nextTokenIsMove(token token.TokenKind) bool {
	if self.nextToken == nil {
		return false
	}

	if self.nextTokenIs(token) {
		self.next()
		return true
	}

	return false
}

// getAnnotationsAndComments returns every annotation or comment statement that has been read until this call
func (self *Parser) getAnnotationsAndComments(from int) []annotationOrComment {
	result := make([]annotationOrComment, 0)

	previous := from
	self.annotationsOrComments.Reverse(func(key int, value *[]annotationOrComment) bool {
		if key < from && previous-key <= 1 {
			result = append(result, *value...)
			previous = key
		}

		return true
	})

	self.resetAnnotationCommentsMap()

	if len(result) == 0 {
		return nil
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

func (self *Parser) pushAnnotation(line int, stmt *AnnotationStmt) {
	value, ok := self.annotationsOrComments.GetMut(line)
	if !ok {
		value = &[]annotationOrComment{{annotation: stmt}}
		self.annotationsOrComments.Set(line, value)
	} else {
		(*value) = append((*value), annotationOrComment{annotation: stmt})
	}
}

func (self *Parser) pushComment(line int, stmt *CommentStmt) {
	value, ok := self.annotationsOrComments.GetMut(line)
	if !ok {
		value = &[]annotationOrComment{{comment: stmt}}
		self.annotationsOrComments.Set(line, value)
	} else {
		(*value) = append((*value), annotationOrComment{comment: stmt})
	}
}

func (self *Parser) resetAnnotationCommentsMap() {
	self.annotationsOrComments = new(btree.Map[int, *[]annotationOrComment])
}

func (self *Parser) reportExpectedNextTokenErr(expected token.TokenKind) {
	var nextToken token.Token
	var pos *reference.Pos
	if self.nextToken == nil {
		nextToken = *token.Token_EOF
		pos = self.tokenizer.GetCurrentPosition()
	} else {
		nextToken = *self.nextToken.token
		pos = self.nextToken.position
	}

	if nextToken.IsEOF() {
		self.pushError(ErrUnexpectedEOF{}, pos)
	} else {
		self.pushError(ErrUnexpectedToken{expected, nextToken}, pos)
	}
}

func (self *Parser) reportExpectedCurrentTokenErr(expected token.TokenKind) {
	var currentToken token.Token
	var pos *reference.Pos
	if self.currentToken == nil {
		currentToken = *token.Token_EOF
		pos = self.tokenizer.GetCurrentPosition()
	} else {
		currentToken = *self.currentToken.token
		pos = self.currentToken.position
	}

	if currentToken.IsEOF() {
		self.pushError(ErrUnexpectedEOF{}, pos)
	} else {
		self.pushError(ErrUnexpectedToken{expected, currentToken}, pos)
	}
}

func (self Parser) getReference(values ...int) *reference.Reference {
	return reference.NewReference(self.file.Path, reference.NewPos(values...))
}

func (self *Parser) pushError(err ParserErrorKind, pos ...*reference.Pos) {
	var position *reference.Pos
	if len(pos) == 0 {
		position = self.tokenizer.GetCurrentPosition()
	} else {
		position = pos[0]
	}

	self.errors.push(NewParserErr(err, reference.NewReference(self.file.Path, position)))
}
