package internal

import "strings"

// TypeResolver maintains a list of types and their ids, and check imports for every file
type TypeResolver struct {
	tree     *AstTree
	contexts []*importContext
}

type AstTree struct {
	packageName string
	sources     []*Ast
	children    []*AstTree
}

type importContext struct {
	owner    *Ast             // The Ast that imports imported
	imported map[*Ast]*string // The list of Ast imported by owner. Being the value of the map an alias
}

// NewTypeResolver creates a new TypeResolver
func NewTypeResolver(source *AstTree) *TypeResolver {
	typeResolver := &TypeResolver{tree: source, contexts: make([]*importContext, 0)}
	return typeResolver
}

// Resolve resolves all the imports for the given AstTree
func (tr *TypeResolver) Resolve() {
	tr.resolveTree(tr.tree)
}

func (tr *TypeResolver) resolveTree(tree *AstTree) {
	tr.resolveSources(tree.sources)
	for _, child := range tree.children {
		tr.resolveTree(child)
	}
}

func (tr *TypeResolver) resolveSources(sources []*Ast) {
	for _, ast := range sources {

		// without a file its impossible to determime paths
		if ast.file == nil {
			continue
		}

		context := &importContext{owner: ast, imported: make(map[*Ast]*string)}
		if ast.imports == nil {
			continue
		}

		for _, importStmt := range *ast.imports {
			packagePath := importStmt.path.lit
			var alias *string
			if importStmt.alias != nil {
				alias = &importStmt.alias.lit
			}

			// lookup package, if not found, continue with the next
			validSources, ok := tr.tree.Lookup(packagePath)
			if !ok {
				continue
			}

			for _, ast := range validSources {
				context.imported[ast] = alias
			}
		}

		tr.contexts = append(tr.contexts, context)
	}
}

// Lookup iterates over the AstTree and returns the list of Ast that are in the given packageName
func (s *AstTree) Lookup(packageName string) ([]*Ast, bool) {
	folders := strings.Split(packageName, "/")
	return s.lookup(folders)
}

func (s *AstTree) lookup(frags []string) ([]*Ast, bool) {
	if len(frags) == 0 {
		return s.sources, true
	}

	if len(frags) == 1 && s.packageName == frags[0] {
		return s.sources, true
	}

	for _, child := range s.children {
		if child.packageName == frags[0] {
			return child.lookup(frags[1:])
		}
	}

	return nil, false
}
