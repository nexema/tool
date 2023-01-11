package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/karrick/godirwalk"
	"github.com/mitchellh/hashstructure/v2"
	"gopkg.in/yaml.v3"
	"tomasweigenast.com/schema_interpreter/internal"
)

const builderVersion = 1
const nexExtension = ".nex"

// Builder provides a method to build a list of .nex files
type Builder struct {
	rootFolder    string
	currentParser *internal.Parser
	analyzer      *internal.Analyzer

	contexts []*internal.ResolvedContext
	typesId  map[*internal.TypeStmt]string
	imports  map[string][]string // the list of import stmt for each file
	cfg      MPackSchemaConfig

	astList         []*internal.Ast
	builtDefinition *internal.NexemaDefinition
}

// NewBuilder creates a new Builder
func NewBuilder() *Builder {
	return &Builder{
		astList: make([]*internal.Ast, 0),
		imports: make(map[string][]string),
	}
}

// Build is the main entry point for parsing nexema projects. A folder is given and
// it first scan for a nexema.yaml, and then start scanning .nex files.
func (b *Builder) Build(inputFolder string) error {

	b.rootFolder = inputFolder

	// the first step is to search for the nexema.yaml in the root folder. Do not search in subfolders. inputFolder
	// should be the root package.
	err := b.scanProject()
	if err != nil {
		return err
	}

	// now, start walking directories
	err = godirwalk.Walk(inputFolder, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if de.IsDir() {
				return nil // skip
			}

			if filepath.Ext(osPathname) != nexExtension {
				return godirwalk.SkipThis
			}

			// scan file
			return b.scanFile(osPathname)
		},
		Unsorted:            true,
		FollowSymbolicLinks: false,
		AllowNonDirectory:   false,
	})

	if err != nil {
		return err
	}

	// build the AstTree
	astTree := internal.NewAstTree(b.astList)

	// resolve types
	resolvedContextArr := internal.NewTypeResolver(astTree).Resolve()

	// analyze the resolved context array
	b.analyzer = internal.NewAnalyzer(resolvedContextArr)
	resolvedContextArr, typesId, errors := b.analyzer.AnalyzeSyntax()
	if !errors.IsEmpty() {
		return errors.Format()
	}

	b.contexts = resolvedContextArr
	b.typesId = typesId

	// now build the definition
	definition := b.buildDefinition()

	// store it
	b.builtDefinition = definition
	return nil
}

// Generates generates source code for each generator specified in cfg.Generators.
// This method must be called after b.Build
func (b *Builder) Generate() error {
	if b.builtDefinition == nil {
		return errors.New("definition not build")
	}

	// serialize definition
	buf, err := json.Marshal(b.builtDefinition)
	if err != nil {
		return err
	}

	// create plugin for each generator
	for generatorName, generator := range b.cfg.Generators {
		if generator.BinPath != "" {
			generator.BinPath = generatorName
		}

		// create plugin
		plugin := NewPlugin(generatorName, generator.BinPath)
		err := plugin.Run(buf)
		if err != nil {
			return err
		}
	}

	return nil
}

// Snapshot generates and saves a snapshot for b.builtDefinition.
// This method must be called after b.Build
func (b *Builder) Snapshot(outFolder string) error {
	if b.builtDefinition == nil {
		return errors.New("definition not build")
	}

	outPath := filepath.Join(outFolder, fmt.Sprintf("%d.nexs", b.builtDefinition.Hashcode))

	err := os.Mkdir(filepath.Dir(outPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not save snapshot. %s", err.Error())
	}

	// todo: serialize to binary using nexemab (nexema binary)
	buf, _ := json.Marshal(b.builtDefinition)
	err = os.WriteFile(outPath, buf, os.ModePerm)

	if err != nil {
		return fmt.Errorf("could not save snapshot. %s", err.Error())
	}

	return nil
}

// Format formats all .nex files in the project.
// This method must be called after b.Build
func (b *Builder) Format() error {
	/*if b.builtDefinition == nil {
		return errors.New("definition not build")
	}

	sb := new(strings.Builder)
	for _, file := range b.builtDefinition.Files {
		sb.Reset()

		// write imports
		imports := b.imports[file.Name]
		if len(imports) > 0 {
			sb.WriteString("import:\n")

			for _, imp := range imports {
				sb.WriteString(fmt.Sprintf("\t%s\n", imp))
			}
		}

		// write types
		for _, typeDef := range file.Types {
			sb.WriteString(fmt.Sprintf("type %s %s {\n", typeDef.Name, typeDef.Modifier))

			// first run is to calculate lengths
			maxIndexLength := 1
			maxNameLength := 1
			maxTypeNameLength := 1
			for _, field := range typeDef.Fields {
				idxLength := len(fmt.Sprint(field.Index))
				if idxLength > maxIndexLength {
					maxIndexLength = idxLength
				}

				nameLength := len(field.Name)
				if nameLength > maxNameLength {
					maxNameLength = nameLength
				}

				switch v := field.Type.(type) {
				case internal.NexemaPrimitiveValueType:
					maxTypeNameLength = len(v.Primitive)
				}
			}

			for _, field := range typeDef.Fields {
				sb.WriteString(fmt.Sprint(field.Index))

				sb.WriteRune('\n')
			}

			sb.WriteString("}\n\n")
		}

		// re-write src
		src := sb.String()
		println(src)
		// err := os.WriteFile(file.Name, []byte(src), os.ModePerm)
		// if err != nil {
		// 	return fmt.Errorf("could not rewrite file. %s", err.Error())
		// }
	}*/

	return nil
}

// scanProject searches and parses a nexema.yaml file in the current folder.
// If the file cannot be found, an error is returned.
func (b *Builder) scanProject() error {
	buf, err := os.ReadFile(filepath.Join(b.rootFolder, "nexema.yaml"))
	if err != nil {
		return fmt.Errorf("nexema.yaml could not be read. Error: %s", err.Error())
	}

	err = yaml.Unmarshal(buf, &b.cfg)
	if err != nil {
		return fmt.Errorf("invalid nexema.yaml file. Error: %s", err.Error())
	}

	return nil
}

// scanFile scans the given file in order to build an Ast. If success, appends the Ast to b.astList,
// otherwise, an error is returned.
func (b *Builder) scanFile(path string) error {
	fileContents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not scan file %s. Error: %s", path, err)
	}

	// todo: re-use the parser
	b.currentParser = internal.NewParser(bytes.NewBuffer(fileContents))

	// parse
	ast, err := b.currentParser.Parse()
	if err != nil {
		return err
	}

	// set file
	relPath, _ := filepath.Rel(b.rootFolder, path)
	ast.File = &internal.File{
		Name: filepath.Base(path),
		Pkg:  filepath.Dir(relPath),
	}

	// append list
	b.astList = append(b.astList, ast)

	return nil
}

// buildDefinition takes the analyzed source and builds a NexemaDefinition for the entire project.
func (b *Builder) buildDefinition() *internal.NexemaDefinition {
	def := &internal.NexemaDefinition{
		Version:  builderVersion,
		Hashcode: 0,
		Files:    make([]internal.NexemaFile, 0),
	}

	files := map[string]*internal.NexemaFile{}

	for _, ctx := range b.contexts {
		ast := ctx.Owner
		fpath := filepath.Join(ast.File.Pkg, ast.File.Name)
		nexemaFile, ok := files[fpath]
		if !ok {
			nexemaFile = &internal.NexemaFile{
				Name:  fpath,
				Types: make([]internal.NexemaTypeDefinition, 0),
			}
			files[fpath] = nexemaFile
		}

		b.imports[fpath] = []string{}
		for _, importStmt := range *ast.Imports {
			b.imports[fpath] = append(b.imports[fpath], fmt.Sprintf("%s as %s", importStmt.Path.Lit, importStmt.Alias.Lit))
		}

		for _, typeStmt := range *ast.Types {
			typeId := b.typesId[typeStmt]
			typeDefinition := internal.NexemaTypeDefinition{
				Id:            typeId,
				Name:          typeStmt.Name.Lit,
				Modifier:      typeStmt.Modifier.String(),
				Documentation: make([]string, 0),
				Fields:        make([]internal.NexemaTypeFieldDefinition, 0),
			}

			if typeStmt.Documentation != nil {
				for _, stmt := range *typeStmt.Documentation {
					typeDefinition.Documentation = append(typeDefinition.Documentation, stmt.Text)
				}
			}

			if typeStmt.Fields != nil {
				for _, stmt := range *typeStmt.Fields {
					field := internal.NexemaTypeFieldDefinition{
						Index:    (stmt.Index.(*internal.PrimitiveValueStmt)).RawValue.(int64),
						Name:     stmt.Name.Lit,
						Metadata: make(map[string]any),
					}

					if typeStmt.Modifier != internal.Token_Enum {
						primitive := internal.GetPrimitive(stmt.ValueType.Ident.Lit)
						switch primitive {
						case internal.Primitive_Illegal, internal.Primitive_Type:
							// Get type id
							var alias *string
							if stmt.ValueType.Ident.Alias != "" {
								alias = &stmt.ValueType.Ident.Alias
							}

							t, _ := ctx.LookupType(stmt.ValueType.Ident.Lit, alias)
							id := b.typesId[t]

							valueType := internal.NexemaTypeValueType{
								BaseNexemaValueType: internal.BaseNexemaValueType{
									Kind:     "NexemaTypeValueType",
									Nullable: stmt.ValueType.Nullable,
								},
								TypeId:      id,
								ImportAlias: alias,
							}
							field.Type = valueType

						default:
							valueType := internal.NexemaPrimitiveValueType{
								BaseNexemaValueType: internal.BaseNexemaValueType{
									Kind:     "NexemaPrimitiveValueType",
									Nullable: stmt.ValueType.Nullable,
								},
								Primitive:     primitive.String(),
								TypeArguments: make([]internal.NexemaValueType, 0),
							}

							if stmt.ValueType.TypeArguments != nil {
								for _, typeArg := range *stmt.ValueType.TypeArguments {
									valueType.TypeArguments = append(valueType.TypeArguments, internal.NexemaPrimitiveValueType{
										BaseNexemaValueType: internal.BaseNexemaValueType{
											Kind:     "NexemaPrimitiveValueType",
											Nullable: typeArg.Nullable,
										},
										Primitive:     internal.GetPrimitive(typeArg.Ident.Lit).String(),
										TypeArguments: make([]internal.NexemaValueType, 0),
									})
								}
							}

							field.Type = valueType
						}

						if stmt.DefaultValue != nil {
							field.DefaultValue = stmt.DefaultValue.Value()
						}
					}

					if stmt.Metadata != nil {
						for _, entry := range *stmt.Metadata {
							key := (entry.Key.(*internal.PrimitiveValueStmt)).RawValue.(string)
							value := (entry.Value.(*internal.PrimitiveValueStmt)).RawValue
							field.Metadata[key] = value
						}
					}

					typeDefinition.Fields = append(typeDefinition.Fields, field)
				}
			}

			nexemaFile.Types = append(nexemaFile.Types, typeDefinition)
		}
	}

	for _, file := range files {
		def.Files = append(def.Files, *file)
	}

	// calculate hashcode
	hash, err := hashstructure.Hash(def.Files, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}

	def.Hashcode = hash

	return def
}
