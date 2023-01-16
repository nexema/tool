package internal

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrNeedAlias    error = errors.New("import needs alias")
	ErrTypeNotFound error = errors.New("type not found")
)

// TypeResolver maintains a list of types and their ids, and check imports for every file
type TypeResolver struct {
	tree     *AstTree
	contexts []*importContext
	errors   []error
}

type AstTree struct {
	packageName string
	sources     []*Ast
	children    []*AstTree
}

type importContext struct {
	owner    *Ast             // The Ast that imports imported
	imported map[*Ast]*string // The list of Ast imported by owner. Being the value of the map an alias
	aliases  map[string]bool  // a map only used to check if an alias is already in use
}

// ResolvedContext maintains a structure where an Ast is the owner of a list of imported Ast
type ResolvedContext struct {
	Owner        *Ast                  // the current Ast which imports another ASTs
	Dependencies map[string][]struct { // the list of packages imported by owner
		Source *Ast    // the imported ast
		Alias  *string // the given import alias
	}
}

// NewAstTree builds an AstTree from a list of Ast
func NewAstTree(astList []*Ast) *AstTree {
	root := &AstTree{
		sources:     make([]*Ast, 0),
		children:    make([]*AstTree, 0),
		packageName: ".",
	}
	for _, ast := range astList {
		file := ast.File
		if file.Pkg == "." { // root folder
			root.packageName = "."
			root.sources = append(root.sources, ast)
		} else {
			// nested tree
			root.append(ast, strings.Split(ast.File.Pkg, "/"))
		}
	}

	return root
}

// NewTypeResolver creates a new TypeResolver
func NewTypeResolver(source *AstTree) *TypeResolver {
	typeResolver := &TypeResolver{tree: source, contexts: make([]*importContext, 0)}
	return typeResolver
}

// Resolve resolves all the imports for the given AstTree.
// If any error is found, its reported
func (tr *TypeResolver) Resolve() []*ResolvedContext {
	tr.resolveTree(tr.tree)

	resolvedContextList := make([]*ResolvedContext, 0, len(tr.contexts))
	for _, context := range tr.contexts {
		resolvedContext := &ResolvedContext{Owner: context.owner, Dependencies: make(map[string][]struct {
			Source *Ast
			Alias  *string
		})}

		for ast, alias := range context.imported {
			_, ok := resolvedContext.Dependencies[ast.File.Pkg]
			if !ok {
				resolvedContext.Dependencies[ast.File.Pkg] = make([]struct {
					Source *Ast
					Alias  *string
				}, 0)
			}

			resolvedContext.Dependencies[ast.File.Pkg] = append(resolvedContext.Dependencies[ast.File.Pkg], struct {
				Source *Ast
				Alias  *string
			}{
				Source: ast,
				Alias:  alias,
			})
		}

		resolvedContextList = append(resolvedContextList, resolvedContext)
	}

	return resolvedContextList
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
		if ast.File == nil {
			continue
		}

		context := &importContext{owner: ast, imported: make(map[*Ast]*string), aliases: make(map[string]bool)}
		if ast.Imports == nil {
			continue
		}

		for _, importStmt := range *ast.Imports {
			packagePath := importStmt.Path.Lit
			var alias *string
			if importStmt.Alias != nil {
				alias = &importStmt.Alias.Lit

				// validate if alias is available
				_, ok := context.aliases[importStmt.Alias.Lit]
				if ok {
					tr.errors = append(tr.errors, fmt.Errorf("duplicated alias %s", importStmt.Alias.Lit))
				} else {
					context.aliases[importStmt.Alias.Lit] = true
				}
			}

			// lookup package, if not found, continue with the next
			validSources, ok := tr.tree.Lookup(packagePath)
			if !ok {
				tr.errors = append(tr.errors, fmt.Errorf("package %s not found", packagePath))
				continue
			}

			for _, ast := range validSources {
				// check for circular dependency
				if tr.isCircular(ast, context) {
					tr.errors = append(tr.errors, fmt.Errorf("circular dependency between %s and %s not allowed", ast.File.Pkg, context.owner.File.Pkg))
					continue
				}
				context.imported[ast] = alias
			}
		}

		tr.contexts = append(tr.contexts, context)
	}
}

// isCircular returns true if the given checkContext.owner is a dependency of ast
func (tr *TypeResolver) isCircular(ast *Ast, checkContext *importContext) bool {
	for _, context := range tr.contexts {
		// first check if the given ast is owner is any context
		if context.owner == ast {
			// if it's, check if the owner of checkContext is in the dependency tree of context
			_, ok := context.imported[checkContext.owner]
			return ok
		}
	}

	return false
}

// append adds a new Ast to s
func (s *AstTree) append(ast *Ast, frags []string) {
	if len(frags) == 0 {
		frags = strings.Split(ast.File.Pkg, "/")
		// frags = frags[len(frags)-1:]
	}

	if len(frags) == 1 {
		for _, child := range s.children {
			if child.packageName == frags[0] {
				child.sources = append(child.sources, ast)
				return
			}
		}

		// if not found, add new children
		s.children = append(s.children, &AstTree{
			packageName: frags[0],
			sources:     []*Ast{ast},
			children:    make([]*AstTree, 0),
		})
		return
	}

	path := frags[0]
	for _, child := range s.children {
		if child.packageName == path {
			child.append(ast, frags[1:])
			return
		}
	}

	// not found, add new children
	node := &AstTree{
		packageName: path,
		sources:     make([]*Ast, 0),
		children:    make([]*AstTree, 0),
	}
	node.append(ast, frags[1:])
	s.children = append(s.children, node)
}

// print prints the tree in a readable way
func (s *AstTree) print(tab string) {
	fmt.Printf("%s- %s\n", tab, s.packageName)
	for _, child := range s.children {
		child.print(tab + " ")
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

// LookupType looks up a TypeStmt named typeName in the current context. If more than one type
// exists with the same name, use alias as tie break.
func (c *ResolvedContext) LookupType(typeName string, alias *string) (*TypeStmt, error) {
	if alias == nil {
		alias = String("")
	}

	// lookup in owner first
	candidates := make(map[*TypeStmt]*string, 0)
	count := 0
	for _, typeStmt := range *c.Owner.Types {
		if typeStmt.Name.Lit == typeName {
			candidates[typeStmt] = nil
			count++
		}
	}

	// lookup in dependencies
	for _, pkg := range c.Dependencies {
		for _, pkgFile := range pkg {
			for _, typeStmt := range *pkgFile.Source.Types {
				if typeStmt.Name.Lit == typeName {
					if pkgFile.Alias != nil && *pkgFile.Alias != *alias {
						continue
					}

					candidates[typeStmt] = pkgFile.Alias
					count++
				}
			}
		}
	}

	// now select from candidates
	if count == 0 {
		return nil, ErrTypeNotFound
	} else if count == 1 {
		for stmt := range candidates {
			return stmt, nil
		}

		return nil, ErrTypeNotFound
	} else {
		if alias == nil {
			return nil, ErrNeedAlias
		}

		for stmt, cantidateAlias := range candidates {
			if *cantidateAlias == *alias {
				return stmt, nil
			}
		}

		return nil, ErrTypeNotFound
	}
}
