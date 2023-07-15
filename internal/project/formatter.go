package project

import (
	"fmt"
	"strings"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/token"
)

type FormatterContext struct {
	parseTree *parser.ParseTree
}

func (self *FormatterContext) formatNode(pkgName string, node *parser.ParseNode) {
	fmt.Printf("Pkg name: %s Nodepath: %s\n", pkgName, node.Path)

	for _, ast := range node.AstList {
		self.formatAst(ast)
	}

	if node.Children != nil {
		node.Iter(func(pkgName string, node *parser.ParseNode) {
			self.formatNode(pkgName, node)
		})
	}
}

func (self *FormatterContext) formatAst(ast *parser.Ast) {
	fmt.Printf("Ast file: %s\n", ast.File.Path)
	stringBuffer := new(strings.Builder)

	if ast.UseStatements != nil {
		for _, stmt := range ast.UseStatements {
			stringBuffer.WriteString(fmt.Sprintf("use %q", stmt.Token.Literal))

			if stmt.Alias != nil {
				stringBuffer.WriteString(fmt.Sprintf(" as %s", stmt.Alias.Token.Literal))
			}

			stringBuffer.WriteRune('\n')
		}
	}

	if ast.TypeStatements != nil {
		for _, stmt := range ast.TypeStatements {
			stringBuffer.WriteString(fmt.Sprintf("type %s", stmt.Name.Token.Literal))
			if stmt.BaseType != nil {
				name, alias := stmt.BaseType.Format()
				if len(alias) > 0 {
					stringBuffer.WriteString(fmt.Sprintf(" extends %s.%s {", alias, name))
				} else {
					stringBuffer.WriteString(fmt.Sprintf(" extends %s {", name))
				}
			} else {
				stringBuffer.WriteString(fmt.Sprintf(" %s {\n", stmt.Modifier.String()))
			}

			maxFieldNameLength := calculateMaxNameLength(stmt.Fields)

			for _, field := range stmt.Fields {
				docsWritten := false
				if field.Documentation != nil {
					for i, doc := range field.Documentation {
						if i == 0 {
							stringBuffer.WriteRune('\n')
						}
						stringBuffer.WriteString(fmt.Sprintf("\t// %s\n", doc.Token.Literal))
					}

					docsWritten = true
				}

				if field.Annotations != nil {
					for i, annotation := range field.Annotations {
						if i == 0 && !docsWritten {
							stringBuffer.WriteRune('\n')
						}
						stringBuffer.WriteString(fmt.Sprintf("\t# %s\n", formatAssignment(&annotation.Assigment)))
					}
				}

				stringBuffer.WriteRune('\t')
				stringBuffer.WriteString(field.Name.Token.Literal)
				stringBuffer.WriteString(strings.Repeat(" ", calculateTabCount(field.Name.Token.Literal, maxFieldNameLength)))
				stringBuffer.WriteString(formatValueType(field.ValueType))
				stringBuffer.WriteRune('\n')
			}

			if stmt.Defaults != nil {
				stringBuffer.WriteString("\n\tdefaults {\n")
				for _, def := range stmt.Defaults {
					stringBuffer.WriteString(fmt.Sprintf("\t\t%s\n", formatAssignment(&def)))
				}
				stringBuffer.WriteString("\t}\n")
			}

			stringBuffer.WriteString("}")
		}
	}

	fmt.Printf("Ast output:\n%s\n", stringBuffer.String())
}

func formatValueType(valueType *parser.DeclStmt) string {
	argsLen := len(valueType.Args)
	sb := new(strings.Builder)
	name, alias := valueType.Format()
	if len(alias) == 0 {
		sb.WriteString(name)
	} else {
		sb.WriteString(fmt.Sprintf("%s.%s", alias, name))
	}
	if valueType.Args != nil && argsLen > 0 {
		sb.WriteRune('(')
		for i, arg := range valueType.Args {
			sb.WriteString(formatValueType(&arg))
			if i < argsLen-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteRune(')')
	}

	if valueType.Nullable {
		sb.WriteRune('?')
	}

	return sb.String()
}

func formatAssignment(assignment *parser.AssignStmt) string {
	if assignment.Right.Token.Kind == token.Map {
		return fmt.Sprintf("%s = %s", "", "")
	}

	return fmt.Sprintf("%s = %s", assignment.Left.Token.Literal, assignment.Right.Kind.Literal())
}

func calculateMaxNameLength(fields []parser.FieldStmt) int {
	maxLength := 0
	for _, field := range fields {
		if len(field.Name.Token.Literal) > maxLength {
			maxLength = len(field.Name.Token.Literal)
		}
	}
	return maxLength
}

func calculateTabCount(name string, maxNameLength int) int {
	return (maxNameLength - len(name) + 4)
}
