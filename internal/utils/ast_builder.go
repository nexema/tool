package utils

import (
	"fmt"
	"reflect"
	"strings"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/token"
)

func MakeIncludeStatement(path, alias string) *parser.IncludeStatement {
	stmt := &parser.IncludeStatement{
		Token: *token.NewToken(token.Include),
		Path: parser.LiteralStatement{
			Token: *token.NewToken(token.String, path),
			Value: parser.StringLiteral{V: path},
		},
	}

	if len(alias) > 0 {
		stmt.Alias = &parser.IdentifierStatement{
			Token: *token.NewToken(token.Ident, alias),
		}
	}

	return stmt
}

func MakeAnnotationStatement(left string, right any) *parser.AnnotationStatement {
	return &parser.AnnotationStatement{
		Token: *token.NewToken(token.Hash),
		Assignation: &parser.AssignStatement{
			Token: *token.NewToken(token.Assign),
			Identifier: &parser.IdentifierStatement{
				Token: *token.NewToken(token.Ident, left),
			},
			Value: MakeLiteralStatement(right),
		},
	}
}

func MakeLiteralStatement(value any) *parser.LiteralStatement {
	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		return &parser.LiteralStatement{
			Token: *token.NewToken(token.String, value.(string)),
			Value: parser.StringLiteral{V: value.(string)},
		}

	case reflect.Int64:
		return &parser.LiteralStatement{
			Token: *token.NewToken(token.Integer, fmt.Sprint(value)),
			Value: parser.IntLiteral{V: value.(int64)},
		}

	case reflect.Float64, reflect.Float32:
		return &parser.LiteralStatement{
			Token: *token.NewToken(token.Decimal, fmt.Sprint(value)),
			Value: parser.FloatLiteral{V: value.(float64)},
		}

	case reflect.Bool:
		return &parser.LiteralStatement{
			Token: *token.NewToken(token.Ident, fmt.Sprint(value)),
			Value: parser.BooleanLiteral{V: value.(bool)},
		}

	default:
		panic(fmt.Sprintf("value kind %s not supported", kind.String()))
	}
}

func MakeTypeStatement(name string, modifier token.TokenKind, extends string, body []parser.Statement) *parser.TypeStatement {
	stmt := &parser.TypeStatement{
		Name: parser.IdentifierStatement{Token: *token.NewToken(token.Ident, name)},
		Modifier: &parser.IdentifierStatement{
			Token: *token.NewToken(modifier),
		},
		Token: *token.NewToken(token.Type),
		Body: &parser.BlockStatement{
			Statements: body,
			Token:      *token.NewToken(token.Lbrace),
		},
	}

	if len(extends) > 0 {
		stmt.Extends = &parser.ExtendsStatement{Token: *token.NewToken(token.Extends), BaseType: parser.IdentifierStatement{
			Token: *token.NewToken(token.Ident, extends),
		}}
	}

	return stmt
}

func MakeFieldStatement(fieldName string, index int, valueType string, otherArgs ...interface{}) *parser.FieldStatement {
	stmt := &parser.FieldStatement{
		Token: *token.NewToken(token.Ident, fieldName),
	}

	if index >= 0 {
		stmt.Index = MakeLiteralStatement(int64(index))
	}

	if len(valueType) > 0 {
		if strings.Contains(valueType, "varchar") {
			number := strings.ReplaceAll(strings.ReplaceAll(valueType, ")", ""), "varchar(", "")
			stmt.ValueType = &parser.DeclarationStatement{
				Token: *token.NewToken(token.Ident, "varchar"),
				Arguments: []parser.DeclarationStatement{
					{
						Token: *token.NewToken(token.Integer, number),
						Identifier: &parser.IdentifierStatement{
							Token: *token.NewToken(token.Integer, number),
						},
					},
				},
			}
		} else {
			var alias, typeName string
			parts := strings.Split(valueType, ".")
			if len(parts) == 2 {
				alias = parts[0]
				typeName = parts[1]
			} else {
				typeName = parts[0]
			}
			stmt.ValueType = &parser.DeclarationStatement{
				Token: *token.NewToken(token.Ident, typeName),
				Identifier: &parser.IdentifierStatement{
					Token: *token.NewToken(token.Ident, typeName),
				},
			}

			if len(alias) > 0 {
				stmt.ValueType.Identifier.Alias = token.NewToken(token.Ident, alias)
			}
		}

	}

	if len(otherArgs) > 0 && stmt.ValueType != nil {
		nullable := otherArgs[0].(bool)
		stmt.ValueType.Nullable = nullable
	}

	return stmt
}

func MakeCustomFieldStatement(fieldName string, index int, valueType *parser.DeclarationStatement) *parser.FieldStatement {
	stmt := &parser.FieldStatement{
		Token:     *token.NewToken(token.Ident, fieldName),
		ValueType: valueType,
	}

	if index >= 0 {
		stmt.Index = MakeLiteralStatement(int64(index))
	}

	return stmt
}

func MakeCommentStatement(text string) *parser.CommentStatement {
	return &parser.CommentStatement{
		Token: *token.NewToken(token.Comment, text),
	}
}

func MakeCommentMultilineStatement(text string) *parser.CommentStatement {
	return &parser.CommentStatement{
		Token: *token.NewToken(token.CommentMultiline, text),
	}
}
