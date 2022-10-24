package internal

import (
	"fmt"
	"strings"
)

type TypeResolver struct {
}

type resolveType struct {
	typeDefinition *TypeDefinition
	imprt          string
}

// NewTypeResolver creates a new TypeResolver
func NewTypeResolver() *TypeResolver {
	return &TypeResolver{}
}

// ResolveAll resolves all types for a given package collection
func (r *TypeResolver) Resolve(tree *DeclarationTree) error {
	return r.resolve(tree.Children)
}

func (r *TypeResolver) resolve(children *[]*DeclarationTree) (err error) {
	if children == nil {
		return nil
	}

	for _, child := range *children {
		if child.Value.IsPackage() {
			err = r.resolve(child.Children)
		} else {
			err = r.resolveFile(child)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TypeResolver) resolveFile(tree *DeclarationTree) error {

	// first, check if all imports are valid
	importContext := make(map[string]*resolveType)
	node := tree.Value.(*FileDeclarationNode)
	for _, imprt := range node.Imports {
		importNode, err := tree.ReverseLookup(imprt)
		if err != nil {
			return fmt.Errorf("file %s was not found", imprt)
		}

		if importNode.Value.IsPackage() {
			return fmt.Errorf("packages cannot be imported: %s", imprt)
		}

		// remove file from path
		lastSlashIdx := strings.LastIndex(imprt, "/")
		if lastSlashIdx != -1 {
			imprt = imprt[:lastSlashIdx]
		}

		for _, t := range *importNode.Value.(*FileDeclarationNode).Types {

			// Check if there is another type with the same name from another package
			other, found := importContext[t.Name]
			if found {
				return fmt.Errorf("cannot import %s. Type %s is imported from %s too", imprt, other.typeDefinition.Name, other.imprt)
			}

			// add to import context
			importContext[t.Name] = &resolveType{
				typeDefinition: t,
				imprt:          imprt,
			}
		}
	}

	// check imports for every type
	for _, typeDef := range *node.Types {

		fields, ok := typeDef.Fields.([]*StructTypeField)
		if !ok {
			continue
		}

		for _, field := range fields {

			var importPath string

			if field.Type.Primitive == Custom {
				importPath = field.Type.ResolveImport
			} else if field.Type.Primitive == List && field.Type.TypeArguments[0].Primitive == Custom {
				importPath = field.Type.TypeArguments[0].ResolveImport
			} else if field.Type.Primitive == Map && field.Type.TypeArguments[1].Primitive == Custom {
				importPath = field.Type.TypeArguments[1].ResolveImport
			} else {
				continue
			}

			var typeDefinition *TypeDefinition
			var ok bool

			// check if the type is a local type or from a import
			typeDefinition, ok = node.Types.LookupType(importPath)
			if !ok {
				// search from imported file
				resolveType, ok := importContext[importPath]
				if !ok {
					return fmt.Errorf("type %s not found in the current context", importPath)
				}

				typeDefinition = resolveType.typeDefinition
			}

			if typeDefinition.Id == typeDef.Id {
				return fmt.Errorf("circular dependency at type %s located at %s. A type cannot be used itself", typeDef.Name, node.Path())
			}

			if typeDefinition.Name == typeDef.Name {
				return fmt.Errorf("cannot import %s because the type which imports it (%s) has the same name", typeDefinition.Name, typeDef.Name)
			}

			field.Type.ImportId = typeDefinition.Id
		}
	}

	return nil
}
