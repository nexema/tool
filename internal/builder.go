package internal

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Builder struct {
	rootPackage string
	mpack       *MPackSchemaDefinition
	rootPath    string

	parser       *Parser
	typeResolver *TypeResolver
	checker      *TypeChecker
	// validator    *Validator
}

// NewBuilder creates a new type builder
func NewBuilder() *Builder {
	return &Builder{
		parser:  NewParser(),
		checker: NewTypeChecker(),
		// validator:    NewValidator(),
	}
}

func (b *Builder) Build(path string) (result *CompileResult, err error) {
	b.rootPath = path

	// Look for mpack.yaml file
	err = b.loadYAML()
	if err != nil {
		return nil, err
	}

	// Get the root package
	b.rootPackage = strings.Split(b.mpack.ProjectName, "/")[1]

	// Parse directory
	tree, err := b.parser.ParseDirectory(path, b.rootPackage)

	if err != nil {
		return nil, err
	}

	b.typeResolver = NewTypeResolver()

	//validate types
	// err = b.validator.Validate(parseResult.Packages)

	// if err != nil {
	// 	return nil, err
	// }

	// Resolve types
	err = b.typeResolver.Resolve(tree)

	if err != nil {
		return nil, err
	}

	return &CompileResult{
		Declaration:      tree,
		Root:             path,
		RootPackage:      b.rootPackage,
		GeneratorOptions: &map[string]interface{}{},
	}, nil
}

// loadYAML loads the mpack.yaml file for the project
func (b *Builder) loadYAML() error {
	path := b.rootPath + "/mpack.yaml"
	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not found nor open mpack.yaml file")
	}

	var schema MPackSchemaDefinition

	err = yaml.Unmarshal(buf, &schema)
	if err != nil {
		return err
	}

	schema.MPackPath = path
	b.mpack = &schema
	return nil
}
