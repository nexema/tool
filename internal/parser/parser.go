package parser

import (
	"io"
	"strconv"

	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
	"tomasweigenast.com/nexema/tool/internal/tokenizer"
)

// Parser is the basic struct for parsing .nex files.
//
// Nexema file structure is defined in the DEFINITION.md file
type Parser struct {
	tokenizer    *tokenizer.Tokenizer
	file         reference.File
	currentToken *tokenBuf
	peekToken    *tokenBuf
	errors       []ParserError
	insideA      token.TokenKind
}

type tokenBuf struct {
	token    *token.Token
	position reference.Pos
}

// NewParser creates a new Parser that has the ability to parse a single source code file.
func NewParser(input io.Reader, file reference.File) *Parser {
	parser := &Parser{
		file:         file,
		currentToken: nil,
		peekToken:    nil,
		tokenizer:    tokenizer.NewTokenizer(input),
		errors:       make([]ParserError, 0),
		insideA:      token.Illegal,
	}

	// feed currentToken and nextToken
	parser.nextToken()
	parser.nextToken()

	return parser
}

func (p *Parser) Parse() *Ast {
	ast := &Ast{
		File:       p.file,
		Statements: make([]Statement, 0),
	}

	for !p.currTokenIs(token.EOF) {
		statement := p.parseStatement()
		if statement != nil {
			ast.Statements = append(ast.Statements, statement)
		}
		p.nextToken()
	}

	return ast
}

func (p *Parser) Errors() []ParserError {
	return p.errors
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	peek, pos, err := p.tokenizer.Next()
	if err != nil {
		p.err(TokenizerErrKind{err.Error()})
	}

	p.peekToken = &tokenBuf{
		token:    peek,
		position: pos,
	}
}

func (p *Parser) parseStatement() Statement {
	switch p.currentToken.token.Kind {
	case token.Include:
		return p.parseIncludeStatement()

	case token.Type:
		return p.parseTypeStatement()

	case token.Integer, token.Ident: // maybe a field
		return p.parseFieldStatement()

	case token.Defaults:
		return p.parseDefaultsStatement()

	case token.Hash:
		return p.parseAnnotationStatement()

	case token.Comment, token.CommentMultiline:
		return p.parseCommentStatement()

	default:
		return nil
	}
}

func (p *Parser) parseCommentStatement() *CommentStatement {
	stmt := &CommentStatement{Token: *p.currentToken.token}
	return stmt
}

func (p *Parser) parseAnnotationStatement() *AnnotationStatement {
	stmt := &AnnotationStatement{Token: *token.NewToken(token.Hash)}
	stmt.Assignation = p.parseAssignStatement()
	return stmt
}

func (p *Parser) parseAssignStatement() *AssignStatement {
	stmt := &AssignStatement{Token: *token.NewToken(token.Assign)}

	if !p.expectPeek(token.Ident) {
		return nil
	}

	stmt.Identifier = p.parseIdentifierStatement()

	if !p.expectPeek(token.Assign) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseLiteralStatement()

	return stmt
}

func (p *Parser) parseIncludeStatement() *IncludeStatement {
	stmt := &IncludeStatement{Token: *p.currentToken.token}

	if !p.expectPeek(token.String) {
		return nil
	}

	stmt.Path = LiteralStatement{
		Token: *p.currentToken.token,
		Value: StringLiteral{value: p.currentToken.token.Literal},
	}

	if p.peekTokenIs(token.As) {
		p.nextToken()
		if p.expectPeek(token.Ident) {
			stmt.Alias = &IdentifierStatement{Token: *p.currentToken.token}
		}
	}

	return stmt
}

func (p *Parser) parseTypeStatement() *TypeStatement {
	stmt := &TypeStatement{Token: *p.currentToken.token}

	if !p.expectPeek(token.Ident) {
		return nil
	}

	stmt.Name = IdentifierStatement{Token: *p.currentToken.token}

	if p.peekTokenIs(token.Extends) {
		p.nextToken()
		stmt.Extends = p.parseExtendsStatement()
	} else if p.peekTokenIs(token.Struct) || p.peekTokenIs(token.Base) || p.peekTokenIs(token.Enum) || p.peekTokenIs(token.Union) {
		p.nextToken()
		stmt.Modifier = &IdentifierStatement{Token: *p.currentToken.token}
	}

	if stmt.Modifier == nil {
		p.insideA = token.Struct
	} else {
		p.insideA = stmt.Modifier.Token.Kind
	}

	if !p.expectPeek(token.Lbrace) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseBlockStatement() *BlockStatement {
	stmt := &BlockStatement{
		Token:      *p.currentToken.token,
		Statements: make([]Statement, 0),
	}

	p.nextToken()
	for !p.currTokenIs(token.Rbrace) && !p.currTokenIs(token.EOF) {
		statement := p.parseStatement()
		if statement != nil {
			stmt.Statements = append(stmt.Statements, statement)
		}

		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseFieldStatement() *FieldStatement {
	stmt := &FieldStatement{}

	if p.currTokenIs(token.Integer) {
		stmt.Index = p.parseLiteralStatement()
		p.nextToken()
	}

	if !p.currTokenIs(token.Ident) {
		return nil
	}

	stmt.Token = *p.currentToken.token

	// for enums return earlier since no more tokens are needed to construct a field
	if p.isInsideA(token.Enum) {
		return stmt
	}

	if !p.expectCurr(token.Ident) {
		return nil
	}

	stmt.ValueType = p.parseDeclarationStatement()

	return stmt
}

func (p *Parser) parseDeclarationStatement() *DeclarationStatement {
	stmt := &DeclarationStatement{}

	if !p.currTokenIs(token.Ident) && !p.currTokenIs(token.Integer) {
		return nil
	}

	stmt.Identifier = p.parseIdentifierStatement()
	stmt.Token = stmt.Identifier.Token

	if p.peekTokenIs(token.Lparen) {
		stmt.Arguments = make([]DeclarationStatement, 0)
		p.nextToken()
		for !p.currTokenIs(token.Rparen) && !p.currTokenIs(token.EOF) {
			p.nextToken()
			argument := p.parseDeclarationStatement()
			if argument == nil {
				return nil
			}

			stmt.Arguments = append(stmt.Arguments, *argument)

			// require comma if more tokens to read
			if !p.peekTokenIs(token.Comma) && !p.peekTokenIs(token.Rparen) {
				p.unexpectedManyError(token.Comma, token.Rparen)
				break
			}

			p.nextToken()
		}
	}

	if p.peekTokenIs(token.QuestionMark) {
		stmt.Nullable = true
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseDefaultsStatement() *DefaultsStatement {
	stmt := &DefaultsStatement{Token: *token.NewToken(token.Defaults)}
	if !p.expectPeek(token.Lbrace) {
		return nil
	}
	stmt.Values = p.parseLiteralStatement()

	return stmt
}

func (p *Parser) parseLiteralStatement() *LiteralStatement {
	stmt := &LiteralStatement{
		Token: *p.currentToken.token,
	}

	stmt.Value = p.parseLiteral(p.currentToken.token)

	return stmt
}

func (p *Parser) parseLiteral(t *token.Token) Literal {
	literal := t.Literal
	switch t.Kind {
	case token.String:
		return StringLiteral{literal}

	case token.Integer:
		value, err := strconv.ParseInt(literal, 10, 64)
		if err != nil {
			panic(err) // maybe a better error handling? is really needed?
		}
		return IntLiteral{value}

	case token.Decimal:
		value, err := strconv.ParseFloat(literal, 64)
		if err != nil {
			panic(err)
		}

		return FloatLiteral{value}

	case token.Ident:
		var value bool
		if literal == "true" {
			value = true
		} else if literal == "false" {
			value = false
		} else {
			return nil
		}

		return BooleanLiteral{value}

	case token.Lbrack:
		listLiteral := ListLiteral{}

		for !p.currTokenIs(token.Rbrack) && !p.currTokenIs(token.EOF) {
			p.nextToken()
			literal := p.parseLiteral(p.currentToken.token)
			if literal == nil {
				break
			}

			listLiteral = append(listLiteral, literal)

			// require comma if more tokens to read
			if !p.peekTokenIs(token.Comma) && !p.peekTokenIs(token.Rbrack) {
				p.unexpectedManyError(token.Comma, token.Rbrack)
				break
			}

			p.nextToken()
		}

		return listLiteral

	case token.Lbrace:
		mapLiteral := MapLiteral{}

		for !p.currTokenIs(token.Rbrace) && !p.currTokenIs(token.EOF) {
			p.nextToken()
			key := p.parseLiteral(p.currentToken.token)
			if key == nil {
				break
			}

			if !p.expectPeek(token.Colon) {
				break
			}

			p.nextToken()

			value := p.parseLiteral(p.currentToken.token)
			if value == nil {
				break
			}

			mapLiteral = append(mapLiteral, MapEntry{key, value})

			// require comma if more tokens to read
			if !p.peekTokenIs(token.Comma) && !p.peekTokenIs(token.Rbrace) {
				p.unexpectedManyError(token.Comma, token.Rbrace)
				break
			}

			p.nextToken()
		}

		return mapLiteral

	default:
		panic("not a valid literal") // todo: add better error handling

	}
}

// parseExtendsStatement parses an statement in the form "extends [identifier]"
func (p *Parser) parseExtendsStatement() *ExtendsStatement {
	stmt := &ExtendsStatement{Token: *p.currentToken.token}
	if !p.expectPeek(token.Ident) {
		return nil
	}

	identifier := p.parseIdentifierStatement()
	if identifier == nil {
		return nil
	}

	stmt.BaseType = *identifier

	return stmt
}

// parseIdentifierStatement parses an statement in the form "(alias.)[identifier]"
func (p *Parser) parseIdentifierStatement() *IdentifierStatement {
	stmt := &IdentifierStatement{}
	ident := p.currentToken.token

	// has an alias
	if p.peekTokenIs(token.Period) {
		p.nextToken()
		if !p.expectPeek(token.Ident) {
			return nil
		}

		stmt.Token = *p.currentToken.token
		stmt.Alias = ident
	} else {
		stmt.Token = *ident
	}

	return stmt
}

// currTokenIs returns true if p.currentToken is t
func (p *Parser) currTokenIs(t token.TokenKind) bool {
	return p.currentToken.token.Kind == t
}

// peekTokenIs returns true if p.peekToken is t
func (p *Parser) peekTokenIs(t token.TokenKind) bool {
	return p.peekToken.token.Kind == t
}

// isInsideA returns true if p.insideA is t
func (p *Parser) isInsideA(t token.TokenKind) bool {
	return p.insideA == t
}

// expectPeek expects p.peekToken is t and advances to the next token
func (p *Parser) expectPeek(t token.TokenKind) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.unexpectedError(t)
		return false
	}
}

// expectCurr expects p.currentToken is t and advances to the next token
func (p *Parser) expectCurr(t token.TokenKind) bool {
	if p.currTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.unexpectedError(t)
		return false
	}
}

// unexpectedError reports an error that happens when parser expects t but founds another one
func (p *Parser) unexpectedError(t token.TokenKind) {
	p.errors = append(p.errors, ParserError{
		At: p.peekToken.position,
		Kind: UnexpectedTokenErrKind{
			Expected: t,
			Got:      p.peekToken.token.Kind,
		},
	})
}

// unexpectedManyError reports an error that happens when parser expects any of t but founds another one
func (p *Parser) unexpectedManyError(t ...token.TokenKind) {
	p.errors = append(p.errors, ParserError{
		At: p.peekToken.position,
		Kind: UnexpectedTokenExpectManyErrKind{
			Expected: t,
			Got:      p.peekToken.token.Kind,
		},
	})
}

func (p *Parser) err(err ParserErrorKind) {
	p.errors = append(p.errors, ParserError{
		At:   p.peekToken.position,
		Kind: err,
	})
}
