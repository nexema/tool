package analysis

import (
	"fmt"
	"path/filepath"
	"strconv"

	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/token"
)

var validMapKeyTypes = map[definition.ValuePrimitive]bool{
	definition.String:  true,
	definition.Boolean: true,
	definition.Uint:    true,
	definition.Uint8:   true,
	definition.Uint16:  true,
	definition.Uint32:  true,
	definition.Uint64:  true,
	definition.Int:     true,
	definition.Int8:    true,
	definition.Int16:   true,
	definition.Int32:   true,
	definition.Int64:   true,
}

// analyzerContext groups related statements and analyzes them. In this case, an Ast
type analyzerContext struct {
	parent        *SemanticAnalyzer // the parent SemanticAnalyzer
	ast           *parser.Ast       // the ast that is being analyzed
	parentContext context           // the parent context of the statement that is being analyzed

	includeStatementsRead bool                          // a flag that indicates if include statements were read
	commentsForNext       []*parser.CommentStatement    // the list of comments (single line) that will be added to the next object
	annotationsForNext    []*parser.AnnotationStatement // the list of annotations that will be added to the next object
	resolvedImports       []include                     // the list of imported Ast files (value is path to Ast)
	aliases               map[string]bool               // a map to keep track of used aliases
	unresolvedReferences  []unresolvedReference         // a list of unresolved references
}

type unresolvedReference struct {
	source   *parser.DeclarationStatement
	typeName string
	alias    string
}

type include struct {
	path  reference.File
	alias string
}

func (ac *analyzerContext) Statement() parser.Statement {
	return ac.ast
}

func (ac *analyzerContext) analyze() {
	ac.annotationsForNext = make([]*parser.AnnotationStatement, 0)
	ac.commentsForNext = make([]*parser.CommentStatement, 0)

	// iterate over every statement declared in the Ast
	for _, statement := range ac.ast.Statements {
		ac.analyzeStatement(statement)
	}
}

func (ac *analyzerContext) analyzeStatement(statement parser.Statement) {
	switch kind := statement.(type) {
	case *parser.IncludeStatement:
		ac.analyzeIncludeStatement(kind)

	case *parser.TypeStatement: // todo: add ServiceStatement when added
		ac.analyzeTypeStatement(kind)

	case *parser.AnnotationStatement, *parser.CommentStatement:
		ac.includeStatementsRead = true
		if comment, ok := kind.(*parser.CommentStatement); ok && comment.Token.Kind == token.Comment {
			ac.commentsForNext = append(ac.commentsForNext, comment)
		}

		if annotation, ok := kind.(*parser.AnnotationStatement); ok {
			ac.annotationsForNext = append(ac.annotationsForNext, annotation)
		}

	case *parser.FieldStatement:
		ac.analyzeFieldStatement(kind)

	case *parser.DefaultsStatement:
		ac.analyzeDefaultsStatement(kind)
	}
}

func (ac *analyzerContext) analyzeIncludeStatement(statement *parser.IncludeStatement) {
	if ac.includeStatementsRead {
		panic("include statements read") // todo: change to err handle
	}

	importPath := statement.Path.Value.Value().(string)
	var alias string
	if statement.Alias != nil {
		alias = statement.Alias.Token.Literal

		if alias == selfSymbols {
			panic(fmt.Errorf("%s cannot be used as an alias because its a reserved keyword", selfSymbols))
		}

		if _, ok := ac.aliases[alias]; ok {
			panic("alias already in use")
		}

		ac.aliases[alias] = true
	}

	importPath = filepath.Clean(importPath)
	if filepath.IsAbs(importPath) {
		rel, err := filepath.Rel(ac.parent.rootFolder, importPath)
		if err != nil {
			panic(err)
		}

		importPath = rel
	}

	nodepath := filepath.Clean(ac.ast.File.Path)
	ac.parent.dependencies.addDependency(nodepath, importPath)
	ac.resolvedImports = append(ac.resolvedImports, include{
		path:  reference.File{Path: importPath},
		alias: alias,
	})
}

func (ac *analyzerContext) analyzeTypeStatement(statement *parser.TypeStatement) {
	ac.includeStatementsRead = true
	if ac.parentContext != nil {
		panic("type inside type!")
	}

	if statement.Name.Alias != nil {
		panic("type name must not have an alias")
	}

	// push symbol to symbolTable
	ac.parent.symbolTable.push(ac.ast.File, selfSymbols, newTypeSymbol(statement))

	ac.parentContext = &typeContext{
		statement:    statement,
		fieldNames:   make(map[string]bool),
		fieldIndexes: make(map[int64]bool),
	}

	for _, body := range statement.Body.Statements {
		ac.analyzeStatement(body)
	}

	ac.parentContext = nil
}

func (ac *analyzerContext) analyzeFieldStatement(statement *parser.FieldStatement) {
	ac.includeStatementsRead = true
	parentContext, ok := ac.parentContext.(*typeContext)
	if !ok {
		panic("not in a type!")
	}

	if parentContext.fieldsRead {
		panic("cannot declare fields after a defaults statement")
	}

	// check for duplicated field index or subsequent indexes
	var fieldIndex int64
	if statement.Index != nil {
		fieldIndex = statement.Index.Value.Value().(int64)

		// first check fi index is not already in use
		if _, ok := parentContext.fieldIndexes[fieldIndex]; ok {
			panic("duplicated field index")
		}

		// check if field index is subsequent from last one
		if parentContext.nextAvailableFieldIdx != fieldIndex {
			panic("index is not subsequent")
		}
		parentContext.nextAvailableFieldIdx++
	} else {
		// deduce the next index for the field
		fieldIndex = parentContext.nextAvailableFieldIdx
		parentContext.nextAvailableFieldIdx++
	}

	// give the index to the field
	parentContext.fieldIndexes[fieldIndex] = true

	// check for duplicated names
	fieldName := statement.Token.Literal
	if _, ok := parentContext.fieldNames[fieldName]; ok {
		panic("duplicated field name")
	}

	// check for nullable fields on union or enum
	if statement.ValueType != nil && statement.ValueType.Nullable && (parentContext.IsUnion() || parentContext.IsEnum()) {
		panic("nullable fields can only be declared on structs or base types")
	}

	// check for valid types
	if statement.ValueType != nil {
		ac.analyzeDeclarationStatement(statement.ValueType)
	}

	parentContext.fieldNames[fieldName] = true
}

func (ac *analyzerContext) analyzeDefaultsStatement(statement *parser.DefaultsStatement) {
	ac.includeStatementsRead = true
	parentContext, ok := ac.parentContext.(*typeContext)
	if !ok {
		panic("not in a type!")
	}

	if !parentContext.fieldsRead {
		panic("no fields were read")
	}

	if parentContext.defaultsRead {
		panic("duplicated defaults")
	}

	defaults := statement.Values.Value.Value().(map[interface{}]interface{})
	for k := range defaults {
		stringKey, ok := k.(string)
		if !ok {
			panic("not a string key")
		}

		if _, ok := parentContext.fieldNames[stringKey]; !ok {
			panic("field not found")
		}
	}

	parentContext.defaultsRead = true
}

func (ac *analyzerContext) analyzeDeclarationStatement(statement *parser.DeclarationStatement) {

	primitive, ok := definition.ParsePrimitive(statement.TokenLiteral())
	if !ok {
		typeName, alias := statement.Get()
		ac.unresolvedReferences = append(ac.unresolvedReferences, unresolvedReference{
			source:   statement,
			typeName: typeName,
			alias:    alias,
		})
		return
	}

	switch primitive {
	case definition.List:
		// expect exactly one argument
		if len(statement.Arguments) != 1 {
			panic("not one argument needed for list")
		}

		argument := statement.Arguments[0]
		ac.analyzeDeclarationStatement(&argument)

	case definition.Map:
		// expect exactly two arguments
		if len(statement.Arguments) != 2 {
			panic("not two arguments needed for map")
		}

		keyArgument := statement.Arguments[0]
		primitive, ok := definition.ParsePrimitive(keyArgument.TokenLiteral())
		if !ok {
			panic("not valid key")
		}

		if _, ok := validMapKeyTypes[primitive]; !ok {
			panic("primitive not available for key")
		}

	case definition.Varchar:
		// expect exactly one argument
		if len(statement.Arguments) != 1 {
			panic("varchar needs exactly one argument")
		}

		argument := statement.Arguments[0]
		charcount, err := strconv.ParseInt(argument.Identifier.Token.Literal, 10, 64)
		if err != nil {
			panic(fmt.Errorf("cannot parse varchar length: %s", err))
		}

		// varchar not more than 2048 characters
		if charcount > 2048 {
			panic("varchar cannot have more than 2048 characters")
		} else if charcount < 1 {
			panic("varchar cannot be lower than 1 character")
		}
	}
}
