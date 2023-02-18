package parser

import (
	"io"
	"strconv"

	"tomasweigenast.com/nexema/tool/token"
	"tomasweigenast.com/nexema/tool/tokenizer"
	"tomasweigenast.com/nexema/tool/utils"
)

type Parser struct {
	tokenizer             tokenizer.Tokenizer
	file                  *File
	currentToken          *tokenBuf
	nextToken             *tokenBuf
	errors                *ParserErrorCollection
	eof                   bool
	annotationsOrComments *utils.OMap[int, *[]annotationOrComment]
}

type tokenBuf struct {
	token    *token.Token
	position *tokenizer.Pos
}

func NewParser(input io.Reader, file *File) *Parser {
	return &Parser{
		file:                  file,
		eof:                   false,
		currentToken:          nil,
		nextToken:             nil,
		tokenizer:             *tokenizer.NewTokenizer(input),
		errors:                newParserErrorCollection(),
		annotationsOrComments: utils.NewOMap[int, *[]annotationOrComment](),
		// annotationsOrComments: btree.NewG(32, annotationOrCommentEntryComparator),
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
		self.reportErr(ErrUnexpectedValue{"literal with import path", *token.Token_EOF})
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
		/* TODO: probably will need this
				let err = self.errors.pop();
		            if let Some(info) = err {
		                if let ParserErr::UnexpectedEofErr = info.error {
		                    self.report_error(ParserErr::ExpectedLiteral(Token::EOF));
		                }
		            };*/

		return nil
	}

	return &AssignStmt{
		Token: *token.NewToken(token.Assign),
		Left:  *ident,
		Right: *literal,
		Pos:   *tokenizer.NewPos(ident.Pos.Start, literal.Pos.End, ident.Pos.Line, literal.Pos.Endline),
	}
}

// parseIdent parses an identifier.
func (self *Parser) parseIdent() *IdentStmt {
	if self.currentToken == nil {
		self.reportErr(ErrExpectedIdentifier{*token.Token_EOF})
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
			self.reportErr(ErrNumberParse{err, literal})
			return nil
		}

		literalKind = IntLiteral{num}

	case token.Decimal:
		num, err := strconv.ParseFloat(literal, 64)
		if err != nil {
			self.reportErr(ErrNumberParse{err, literal})
			return nil
		}

		literalKind = FloatLiteral{num}

	case token.Ident:
		if literal == "true" {
			literalKind = BooleanLiteral{true}
		} else if literal == "false" {
			literalKind = BooleanLiteral{false}
		} else {
			self.reportErr(ErrInvalidLiteral{literalToken})
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
		if self.expectCurrentToken(token.Rbrack) {
			endPos := self.currentToken.position
			tokenPos = *tokenizer.NewPos(tokenPos.Start, endPos.End, tokenPos.Line, endPos.Endline)
		}

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
	default:
		self.reportErr(ErrInvalidLiteral{literalToken})
		return nil
	}

	return &LiteralStmt{
		Kind: literalKind,
		Pos:  tokenPos,
	}
}

// next reads the next token, skipping comments and saving them for later use
func (self *Parser) next() {
	err := self.consume()
	if err != nil {
		self.errors.push(NewParserErr(ErrTokenizer{}, *self.tokenizer.GetCurrentPosition()))
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
		self.annotationsOrComments.Upsert(line, func(value *[]annotationOrComment) {
			*value = append(*value, annotationOrComment{
				comment: &CommentStmt{
					Token: *currToken.token,
					Pos:   *currToken.position,
				},
			})
		}, newAnnotationOrCommentArray)

		self.next()
		return

	case token.Hash:
		self.next()
		annotationStmt := self.parseAnnotationStmt()
		if annotationStmt != nil {
			line := annotationStmt.Pos.Endline
			self.annotationsOrComments.Upsert(line, func(value *[]annotationOrComment) {
				*value = append(*value, annotationOrComment{
					annotation: annotationStmt,
				})
			}, newAnnotationOrCommentArray)

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
		return true
	}

	self.reportExpectedNextTokenErr(token)
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

func (self *Parser) reportExpectedNextTokenErr(expected token.TokenKind) {
	var nextToken token.Token
	var pos tokenizer.Pos
	if self.nextToken == nil {
		nextToken = *token.Token_EOF
		pos = *self.tokenizer.GetCurrentPosition()
	} else {
		nextToken = *self.nextToken.token
		pos = *self.nextToken.position
	}

	if nextToken.IsEOF() {
		self.errors.push(NewParserErr(ErrUnexpectedEOF{}, pos))
	} else {
		self.errors.push(NewParserErr(ErrUnexpectedToken{expected, nextToken}, pos))
	}
}

func (self *Parser) reportErr(err ParserErrorKind) {
	var pos tokenizer.Pos
	if self.currentToken == nil {
		pos = *self.tokenizer.GetCurrentPosition()
	} else {
		pos = *self.currentToken.position
	}

	self.errors.push(NewParserErr(err, pos))
}
