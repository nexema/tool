package internal

import (
	"path/filepath"

	"github.com/mitchellh/hashstructure/v2"
)

const builderVersion = 1

// Builder provides a method to build a list of .nex files
type Builder struct {
	currentParser *Parser
	analyzer      *Analyzer
}

// NewBuilder creates a new Builder
func NewBuilder() *Builder {
	return &Builder{}
}

// Build is the main entry point for parsing nexema projects. A folder is given and
// it first scan for a nexema.yaml, and then start scanning .nex files.
func (b *Builder) Build() {

}

// buildDefinition takes the analyzed source and builds a NexemaDefinition
func (b *Builder) buildDefinition() *NexemaDefinition {
	def := &NexemaDefinition{
		Version:  builderVersion,
		Hashcode: 0,
		Files:    make([]NexemaFile, 0),
	}

	files := map[string]*NexemaFile{}

	for _, ctx := range b.analyzer.contexts {
		ast := ctx.owner
		fpath := filepath.Join(ast.file.pkg, ast.file.name)
		nexemaFile, ok := files[fpath]
		if !ok {
			nexemaFile = &NexemaFile{
				Name:  fpath,
				Types: make([]NexemaTypeDefinition, 0),
			}
			files[fpath] = nexemaFile
		}

		for _, typeStmt := range *ast.types {
			typeId := b.analyzer.typesId[typeStmt]
			typeDefinition := NexemaTypeDefinition{
				Id:            typeId,
				Name:          typeStmt.name.lit,
				Modifier:      typeStmt.modifier.String(),
				Documentation: make([]string, 0),
				Fields:        make([]NexemaTypeFieldDefinition, 0),
			}

			if typeStmt.documentation != nil {
				for _, stmt := range *typeStmt.documentation {
					typeDefinition.Documentation = append(typeDefinition.Documentation, stmt.text)
				}
			}

			if typeStmt.fields != nil {
				for _, stmt := range *typeStmt.fields {
					field := NexemaTypeFieldDefinition{
						Index:    (stmt.index.(*PrimitiveValueStmt)).value.(int64),
						Name:     stmt.name.lit,
						Metadata: make(map[string]any),
					}

					if typeStmt.modifier != Token_Enum {
						primitive := GetPrimitive(stmt.valueType.ident.lit)
						switch primitive {
						case Primitive_Illegal, Primitive_Type:
							// Get type id
							var alias *string
							if stmt.valueType.ident.alias != "" {
								alias = &stmt.valueType.ident.alias
							}

							t, _ := ctx.lookupType(stmt.valueType.ident.lit, alias)
							id := b.analyzer.typesId[t]

							valueType := NexemaTypeValueType{
								Base: BaseNexemaValueType{
									Kind:     "NexemaTypeValueType",
									Nullable: stmt.valueType.nullable,
								},
								TypeId:      id,
								ImportAlias: alias,
							}
							field.Type = valueType

						default:
							valueType := NexemaPrimitiveValueType{
								Base: BaseNexemaValueType{
									Kind:     "NexemaPrimitiveValueType",
									Nullable: stmt.valueType.nullable,
								},
								Primitive:     primitive.String(),
								TypeArguments: make([]nexemaValueType, 0),
							}

							if stmt.valueType.typeArguments != nil {
								for _, typeArg := range *stmt.valueType.typeArguments {
									valueType.TypeArguments = append(valueType.TypeArguments, NexemaPrimitiveValueType{
										Base: BaseNexemaValueType{
											Kind:     "NexemaPrimitiveValueType",
											Nullable: typeArg.nullable,
										},
										Primitive:     GetPrimitive(typeArg.ident.lit).String(),
										TypeArguments: make([]nexemaValueType, 0),
									})
								}
							}

							field.Type = valueType
						}

						if stmt.defaultValue != nil {
							field.DefaultValue = stmt.defaultValue.Value()
						}
					}

					if stmt.metadata != nil {
						for _, entry := range *stmt.metadata {
							key := (entry.key.(*PrimitiveValueStmt)).value.(string)
							value := (entry.value.(*PrimitiveValueStmt)).value
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
