package internal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var typeNameRegex = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_]*$")

// var packageNameRegex = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_.]*$")
var ErrEndOfType = errors.New("end of type")

type Parser struct {
	structFieldTokenizer *StructFieldTokenizer
	valueParser          *ValueParser

	schema *MPackSchemaDefinition
}

func NewParser(schema ...*MPackSchemaDefinition) *Parser {
	var schemaArg *MPackSchemaDefinition
	if len(schema) > 0 {
		schemaArg = schema[0]
	}

	return &Parser{
		structFieldTokenizer: NewStructFieldTokenizer(),
		valueParser:          NewValueParser(),
		schema:               schemaArg,
	}
}

// ParseDirectory Parses a directory containing .mpack files
// This is the main method to start parsing .mpack structures, it takes a path to the
// root directory, and outputs a result containing all the packages read
func (p *Parser) ParseDirectory(path string, rootPackage string) (result *DeclarationTree, err error) {
	if err != nil {
		return nil, err
	}

	rootPackageName := filepath.Base(path)
	// NewPackageDeclarationNode(rootPackageName, path, make([]string, 0))]
	tree := NewDeclarationTree(rootPackageName, NewPackageDeclarationNode(rootPackageName, path), make([]*DeclarationTree, 0))

	// iterate over files and directories
	err = filepath.Walk(path, func(fPath string, d fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// ignore base package
			if strings.HasSuffix(fPath, rootPackage) {
				return nil
			}

			// Add nested package
			tree.Add(NewPackageDeclarationNode(d.Name(), fPath))

			// get package hierarchy
			// packageHierarchy := Skip(strings.Split(fPath, "\\"), rootPackage)

			// err = packages.CreatePackage(packageHierarchy, fPath)
			// if err != nil {
			// 	return err
			// }

		} else {
			if filepath.Ext(fPath) != ".mpack" {
				return nil
			}

			if p.schema != nil {
				for _, skip := range p.schema.SkipFiles {
					if fPath == skip {
						return nil
					}
				}
			}

			// get dir name
			// dirName := filepath.Dir(fPath)
			// dirName = strings.Join(Skip(strings.Split(dirName, "\\"), rootPackage), ".")
			// packageName := dirName

			// pkg, err := packages.GetPackage(packageName)
			// if err != nil {
			// 	return err
			// }

			// parse a file
			file, err := p.ParseFile(fPath, d.Name())
			if err != nil {
				return err
			}

			// Append node
			tree.Add(NewFileDeclarationNode(file.Id, file.FilePath, file.FileName, file.Imports, file.Types))

			// pkg.Files = append(pkg.Files, file)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tree, nil
}

// ParseFile Parses a single .mpack file
func (p *Parser) ParseFile(path string, fileName string) (result *FileDeclarationNode, err error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	definition, err := p.internalParse(NewReader(scanner), fileName)
	if err != nil {
		return nil, err
	}

	definition.FilePath = path
	return definition, nil
}

// ParseString Parses a string containing a .mpack definition
func (p *Parser) ParseString(s string) (result *FileDeclarationNode, err error) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	return p.internalParse(NewReader(scanner), "")
}

func (p *Parser) internalParse(scanner *Reader, fileName string) (*FileDeclarationNode, error) {
	fileDefinition := &FileDeclarationNode{
		FileName: fileName,
		Imports:  make([]string, 0),
		Types:    make([]*TypeDefinition, 0),
	}

	for {
		token, err := scanner.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fileDefinition, nil
			}

			return nil, err
		}

		token = strings.TrimSpace(token)

		// Skip comments
		if strings.HasPrefix(token, "//") {
			continue
		}

		// Skip empty lines
		if len(token) == 0 {
			continue
		}

		// Read types
		typeDef, err := p.readType(scanner)
		if err != nil {
			return nil, err
		}

		if typeDef != nil {
			fileDefinition.Types = append(fileDefinition.Types, typeDef)
			continue
		} else {
			break
		}
	}

	// fileDefinition.Id = GetTypeHash(fileDefinition.Name, fileDefinition.PackageName)

	return fileDefinition, nil
}

func (p *Parser) readType(reader *Reader) (*TypeDefinition, error) {
	token := reader.Current()

	if !strings.HasPrefix(token, "type") {
		return nil, nil
	}

	token = token[:len(token)-1]

	tokens := make([]string, 0)
	for _, t := range strings.Split(token, " ") {
		if len(t) > 0 {
			tokens = append(tokens, t)
		}
	}

	tokensLen := len(tokens)
	if tokensLen < 2 {
		return nil, fmt.Errorf("invalid type declaration, expected two or more arguments, given: %s", token)
	}

	typeDef := &TypeDefinition{}
	typeName := tokens[1]

	// Check for a valid type name
	matches := typeNameRegex.MatchString(typeName)
	if !matches {
		return nil, fmt.Errorf("the name must match the following regex: ^[a-zA-Z][a-zA-Z0-9]*$, given: %v", typeName)
	}

	typeDef.Name = strings.TrimSpace(typeName)

	if tokensLen == 3 {
		modifier := tokens[2]

		// Check if modifier is valid
		ok, modifierType := ParseModifier(modifier)
		if !ok {
			return nil, fmt.Errorf("invalid type modifier, given: %s", modifier)
		}

		typeDef.Modifier = modifierType
	} else {
		typeDef.Modifier = Struct
	}

	// Read fields
	fields, err := p.readFields(reader, typeDef.Modifier)
	if err != nil {
		if !errors.Is(err, ErrEndOfType) {
			return nil, err
		}
	}

	typeDef.Fields = fields

	return typeDef, nil
}

func (p *Parser) readFields(reader *Reader, modifier TypeModifier) (interface{}, error) {
	switch modifier {
	case Struct, Union:
		return p.readStructUnionFields(reader)

	case Enum:
		return p.readEnumFields(reader)

	default:
		return nil, fmt.Errorf("unimplemented field reader for type modifier %v", modifier)
	}
}

func (p *Parser) readStructUnionFields(reader *Reader) ([]*StructTypeField, error) {
	fields := make([]*StructTypeField, 0)
	for {
		token, err := reader.Next()
		tokenLen := len(token)

		if err != nil {

			// If end of file, return types
			if errors.Is(err, io.EOF) {
				return fields, nil
			}

			return nil, err
		}

		// If current token is empty, continue to the next line
		if tokenLen == 0 {
			continue
		}

		// Check if end of type
		if token[0] == '}' {
			return fields, ErrEndOfType
		}

		tokenizeResult, err := p.structFieldTokenizer.Tokenize(token)
		if err != nil {
			return nil, err
		}

		field, err := p.readStructField(tokenizeResult)

		if err != nil {
			return nil, err
		}

		fields = append(fields, field.ToField())
	}
}

func (p *Parser) readEnumFields(reader *Reader) ([]*EnumTypeField, error) {
	fields := make([]*EnumTypeField, 0)
	for {
		token, err := reader.Next()
		tokenLen := len(token)

		if err != nil {

			// If end of file, return types
			if errors.Is(err, io.EOF) {
				return fields, nil
			}

			return nil, err
		}

		// If current token is empty, continue to the next line
		if tokenLen == 0 {
			continue
		}

		// Check if end of type
		if token[0] == '}' {
			return fields, ErrEndOfType
		}

		tokens := strings.Split(token, " ")
		if len(tokens) != 2 {
			return nil, fmt.Errorf("invalid enum's value, given: %v", token)
		}

		valueName := tokens[0]
		matches := typeNameRegex.MatchString(valueName)
		if !matches {
			return nil, fmt.Errorf("the name must match the following regex: ^[a-zA-Z][a-zA-Z0-9]*$, given: %v", valueName)
		}

		rawIndex := tokens[1]

		// Parse index
		index, err := strconv.ParseUint(rawIndex, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid enum's value index, expected a positive number, given: %v", rawIndex)
		}

		fields = append(fields, &EnumTypeField{
			Name:  valueName,
			Index: uint32(index),
		})
	}
}

func (p *Parser) readStructField(tokenizeResult *StructFieldTokenizerResult) (*structTypeBuilder, error) {
	fieldBuilder := &structTypeBuilder{}

	// First argument must be the name of the field
	fieldBuilder.Name = tokenizeResult.FieldName
	matches := typeNameRegex.MatchString(fieldBuilder.Name)
	if !matches {
		return nil, fmt.Errorf("the field's name must match the following regex: ^[a-zA-Z][a-zA-Z0-9]*$, given: %v", fieldBuilder.Name)
	}

	// Read field's type
	rawFieldType := tokenizeResult.PrimitiveFieldTypeName
	//fmt.Printf("field: %+v\n", rawFieldType)

	packageHierarchy, fieldTypeName := p.readPackageHierarchy(rawFieldType)
	if packageHierarchy != nil {
		rawFieldType = fieldTypeName
	}

	// Read field type
	primitive, nullable, typeArguments, err := p.readFieldType(rawFieldType, tokenizeResult.TypeArguments)
	if err != nil {
		return nil, err
	}

	fieldBuilder.Type = FieldTypeValue{
		Primitive:        primitive,
		Nullable:         nullable,
		TypeArguments:    typeArguments,
		PackageHierarchy: packageHierarchy,
		ResolveImport:    tokenizeResult.PrimitiveFieldTypeName,
	}

	if primitive == Custom {
		fieldBuilder.Type.TypeName = rawFieldType
	}

	// Read field index
	rawFieldIndex := tokenizeResult.FieldIndex
	fieldIndex, err := strconv.ParseUint(rawFieldIndex, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("cannot parse field index, invalid number, given: %s", rawFieldIndex)
	}
	fieldBuilder.Index = uint32(fieldIndex)

	if len(tokenizeResult.DefaultValue) > 0 {
		defaultValue, err := p.readDefaultValue(tokenizeResult.DefaultValue)
		if err != nil {
			return nil, err
		}

		fieldBuilder.DefaultValue = defaultValue
	}

	if len(tokenizeResult.Metadata) > 0 {
		metadata, err := p.readMetadata(tokenizeResult.Metadata)
		if err != nil {
			return nil, err
		}

		fieldBuilder.Metadata = metadata
	}

	return fieldBuilder, nil
}

func (p *Parser) readPackageHierarchy(rawFieldType string) (packageHierarchy []string, typeName string) {
	tokens := strings.Split(rawFieldType, ".")
	tokensLen := len(tokens)
	if tokensLen > 1 {
		return tokens[:tokensLen-1], tokens[tokensLen-1]
	}

	return nil, tokens[0]
}

func (p *Parser) readFieldType(rawFieldType string, typeParams []string) (primitive FieldTypePrimitive, nullable bool, typeArgs []FieldTypeValue, err error) {

	valid, primitive, nullable := ParseFieldType(rawFieldType)
	if !valid {
		return Custom, nullable, nil, nil
	}

	// now read type arguments
	var typeArguments []FieldTypeValue
	if len(typeParams) > 0 {
		paramsLen := len(typeParams)

		isList := paramsLen == 1

		if isList {
			argName := typeParams[0]
			packageHierarchy, argumentName := p.readPackageHierarchy(argName)

			primitive, nullable, _, err := p.readFieldType(argumentName, nil)
			if err != nil {
				return UnknownFieldType, false, nil, err
			}

			typeArguments = append(typeArguments, FieldTypeValue{
				Primitive:        primitive,
				TypeArguments:    nil,
				Nullable:         nullable,
				PackageHierarchy: packageHierarchy,
				ResolveImport:    argName,
			})

		} else {
			keyArgName := typeParams[0]
			valueArgName := typeParams[1]

			// Parse key type
			keyType, nullable, typeArgs, err := p.readFieldType(keyArgName, nil)

			if err != nil {
				return UnknownFieldType, false, nil, err
			}

			// Check if a valid key type
			switch keyType {
			case String, Int8, Int16, Int32, Int64, Uint8, Uint16, Uint32, Uint64, Float32, Float64:
				break

			default:
				return UnknownFieldType, false, nil, fmt.Errorf("a map key only can be of type string, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64, float128")
			}

			if nullable {
				return UnknownFieldType, false, nil, fmt.Errorf("a map key cannot be nullable")
			}

			if typeArgs != nil {
				return UnknownFieldType, false, nil, fmt.Errorf("a map key cannot have type arguments")
			}

			// Parse value type
			valueTypePackageHierarchy, valueArgumentName := p.readPackageHierarchy(valueArgName)

			valueType, nullable, typeArgs, err := p.readFieldType(valueArgumentName, nil)

			if err != nil {
				return UnknownFieldType, false, nil, err
			}

			if typeArgs != nil {
				return UnknownFieldType, false, nil, fmt.Errorf("a map value cannot have type arguments")
			}

			typeArguments = append(typeArguments, FieldTypeValue{
				Primitive:        keyType,
				TypeArguments:    nil,
				Nullable:         false,
				PackageHierarchy: nil,
			})

			typeArguments = append(typeArguments, FieldTypeValue{
				Primitive:        valueType,
				TypeArguments:    nil,
				Nullable:         nullable,
				PackageHierarchy: valueTypePackageHierarchy,
				ResolveImport:    valueArgName,
			})

		}
	}

	return primitive, nullable, typeArguments, nil
}

func (p *Parser) readMetadata(rawMetadata string) (map[string]interface{}, error) {

	if rawMetadata[0] != '[' {
		return nil, fmt.Errorf("expected [ for declaring metadata, given: %s", string(rawMetadata[1]))
	}

	if rawMetadata[len(rawMetadata)-1] != ']' {
		return nil, fmt.Errorf("expected ] for closing metadata map, given: %v", rawMetadata)
	}

	// Skip the []
	rawMetadata = rawMetadata[1 : len(rawMetadata)-1]

	// Split values
	rawMetadataValues := strings.Split(rawMetadata, ",")
	if len(rawMetadataValues) == 0 {
		return nil, errors.New("if you are going to declare a metadata value, you must include at least one value")
	}

	metadata := make(map[string]interface{})
	for _, rawMetadataValue := range rawMetadataValues {
		rawMetadataValue = strings.TrimSpace(rawMetadataValue)

		if rawMetadataValue[0] != '(' {
			return nil, fmt.Errorf("metadata value must contain the open params at the beginning, given: %v", rawMetadataValue)
		}

		if rawMetadataValue[len(rawMetadataValue)-1] != ')' {
			return nil, fmt.Errorf("metadata value must contain the closing params at the end, given: %v", rawMetadataValue)
		}

		// Parse map
		mapValue, parsedType, err := p.valueParser.ParseString(rawMetadataValue)
		if err != nil {
			return nil, err
		}

		if parsedType != Map {
			return nil, errors.New("metadata values must be of type map(string,string|int|boolean)")
		}

		mapR := reflect.TypeOf(mapValue)
		mapKeyR := mapR.Key()
		mapValueR := mapR.Elem()

		if mapKeyR.Kind() != reflect.String {
			return nil, errors.New("metadata keys cannot be of another type than string")
		}

		mapValueRKind := mapValueR.Kind()
		switch mapValueRKind {
		case reflect.String, reflect.Int, reflect.Uint, reflect.Bool:
		default:
			return nil, errors.New("metadata values cannot have other type than string, int or boolean")
		}
	}

	return metadata, nil
}

func (p *Parser) readDefaultValue(rawDefaultValue string) (interface{}, error) {
	v, _, err := p.valueParser.ParseString(rawDefaultValue)
	return v, err
}

type structTypeBuilder struct {
	Name         string
	NameEndIdx   int
	Type         FieldTypeValue
	TypeEndIdx   int
	Index        uint32
	DefaultValue interface{}
	Metadata     map[string]interface{}
}

func (b *structTypeBuilder) ToField() *StructTypeField {
	return &StructTypeField{
		Name:         b.Name,
		Type:         b.Type,
		Index:        b.Index,
		DefaultValue: b.DefaultValue,
		Metadata:     b.Metadata,
	}
}
