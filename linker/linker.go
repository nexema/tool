package linker

import (
	"tomasweigenast.com/nexema/tool/parser"
	"tomasweigenast.com/nexema/tool/scope"
	"tomasweigenast.com/nexema/tool/tokenizer"
	"tomasweigenast.com/nexema/tool/utils"
)

// Linker validates the following:
//
// - Struct names are not duplicated in the current LocalScope
// - Imports points to valid packages
// - Imported types are valid and names does not collide
type Linker struct {
	src    *parser.ParseTree
	scopes []*scope.Scope
	errors *LinkerErrorCollection
}

func NewLinker(parseTree *parser.ParseTree) *Linker {
	return &Linker{
		src:    parseTree,
		scopes: make([]*scope.Scope, 0),
		errors: newLinkerErrorCollection(),
	}
}

func (self *Linker) LinkedScopes() []*scope.Scope {
	return self.scopes
}

func (self *Linker) HasLinkErrors() bool {
	return len(*self.errors) > 0
}

func (self *Linker) Errors() *LinkerErrorCollection {
	return self.errors
}

func (self *Linker) Link() {
	self.buildScopes()
	self.resolveImports()
	self.verifyCircularDependencies()
	self.verifyObjects()
}

// verifyObjects checks if there are no type names which can collide with imported ones, or between them
func (self *Linker) verifyObjects() {
	for _, s := range self.scopes {
		for _, ls := range *s.LocalScopes() {
			m := map[string]*scope.Import{}

			// verify objects between imports
			for resolvedScope, imp := range *ls.ResolvedScopes() {
				for _, obj := range resolvedScope.GetAllObjects() {
					if _, ok := m[obj.Name]; ok {
						// found another object and this import does not has an alias
						if !imp.HasAlias() {
							self.errors.push(NewLinkerErr(ErrAlreadyDefined{obj.Name}, imp.Source().Path.Pos))
							continue
						}
					}

					m[obj.Name] = imp
				}
			}

			// verify local objects against imports
			// objects are not verified against other in local scope because they are already verified at discover stage
			for objName := range *ls.Objects() {
				if imp, ok := m[objName]; ok {
					// if the imported object does not have an alias, report error
					if !imp.HasAlias() {
						self.errors.push(NewLinkerErr(ErrAlreadyDefined{objName}, imp.Source().Path.Pos))
						continue
					}
				}
			}
		}
	}
}

func (self *Linker) verifyCircularDependencies() {
	graph := map[*scope.Scope][]*scope.Scope{}

	// for each scope, add edges to the graph from the current scope
	for _, s := range self.scopes {
		graph[s] = make([]*scope.Scope, 0)
		for _, localScope := range *s.LocalScopes() {
			for resolvedScope := range *localScope.ResolvedScopes() {
				graph[s] = append(graph[s], resolvedScope)
			}
		}
	}

	// perform a topological sort on the graph
	visited := map[*scope.Scope]bool{}
	stack := []*scope.Scope{}

	for _, scope := range self.scopes {
		if _, ok := visited[scope]; !ok {
			src, dest, hasCircular := self.hasCircularDependencies(scope, &graph, &visited, &stack)
			if hasCircular {
				self.errors.push(NewLinkerErr(ErrCircularDependency{src.File(), dest.File()}, *tokenizer.NewPos()))
			}
		}
	}
}

func (self *Linker) hasCircularDependencies(node *scope.Scope, graph *map[*scope.Scope][]*scope.Scope, visited *map[*scope.Scope]bool, stack *[]*scope.Scope) (*scope.LocalScope, *scope.LocalScope, bool) {
	(*visited)[node] = true
	*stack = append(*stack, node)

	for _, neighbor := range (*graph)[node] {
		if _, ok := (*visited)[neighbor]; !ok {
			src, dest, hasCircular := self.hasCircularDependencies(neighbor, graph, visited, stack)
			if hasCircular {
				return src, dest, true
			}
		} else if utils.Contains(stack, neighbor) {
			scope1 := utils.Find(node.LocalScopes(), func(t **scope.LocalScope) bool {
				_, ok := (*(*t).ResolvedScopes())[neighbor]
				return ok
			})

			scope2 := utils.Find(neighbor.LocalScopes(), func(t **scope.LocalScope) bool {
				_, ok := (*(*t).ResolvedScopes())[node]
				return ok
			})

			return *scope1, *scope2, true
		}
	}

	*stack = (*stack)[:len(*stack)-1]

	return nil, nil, false
}

// resolveImports resolves an use statement for every LocalScope
func (self *Linker) resolveImports() {
	for _, pkgScope := range self.scopes {
		for _, localScope := range *pkgScope.LocalScopes() {
			aliases := map[string]bool{}
			for impPath, imp := range *localScope.Imports() {

				// alias already defined
				if imp.HasAlias() {
					if _, ok := aliases[imp.Alias]; ok {
						self.errors.push(NewLinkerErr(ErrAliasAlreadyDefined{imp.Alias}, imp.Source().Alias.Pos))
						continue
					}

					aliases[imp.Alias] = true
				}

				// check if impPath is not equal to pkgScope.Path
				// it would be a self import
				if impPath == pkgScope.Path() {
					self.errors.push(NewLinkerErr(ErrSelfImport{}, imp.Source().Path.Pos))
					continue
				}

				// find scope
				resolvedScope := self.findScope(impPath)
				if resolvedScope == nil {
					self.errors.push(NewLinkerErr(ErrPackageNotFound{impPath}, imp.Source().Path.Pos))
					continue
				}

				localScope.AddResolvedScope(resolvedScope, imp)
			}
		}
	}
}

func (self *Linker) findScope(path string) *scope.Scope {
	for _, scopePkg := range self.scopes {
		if scopePkg.Path() == path {
			return scopePkg
		}
	}

	return nil
}

func (self *Linker) buildScopes() {
	self.src.Root().Iter(func(pkgName string, node *parser.ParseNode) {
		self.createScope(pkgName, node)
	})
}

func (self *Linker) createScope(packageName string, node *parser.ParseNode) {

	newScope := scope.NewScope(node.Path, packageName)
	for _, ast := range node.AstList {
		imports := make(map[string]*scope.Import)
		objects := make(map[string]*scope.Object)

		// validate "use" statements
		for _, stmt := range ast.UseStatements {
			imp := scope.NewImport(&stmt)
			imports[imp.Path] = imp
		}

		// push types
		for _, stmt := range ast.TypeStatements {
			obj := scope.NewObject(&stmt)

			if _, ok := objects[obj.Name]; ok {
				self.errors.push(NewLinkerErr(ErrAlreadyDefined{obj.Name}, obj.Source().Name.Pos))
				continue
			}

			objects[obj.Name] = obj
		}

		newScope.PushLocalScope(scope.NewLocalScope(ast.File, imports, objects))
	}

	self.scopes = append(self.scopes, newScope)

	node.Iter(func(pkgName string, node *parser.ParseNode) {
		self.createScope(pkgName, node)
	})
}
