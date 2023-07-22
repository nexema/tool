package linker

import (
	"path/filepath"

	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// Linker validates the following:
//
// - Struct names are not duplicated in the current LocalScope
// - Imports points to valid packages
// - Imported types are valid and names does not collide
type Linker struct {
	src             *parser.ParseTree
	rootScope       scope.Scope
	dependencyGraph dependencyGraph
	errors          *LinkerErrorCollection
}

type dependencyGraph map[scope.Scope][]scope.Scope
type dependencyConflict struct {
	importer reference.File
	imported reference.File
}

func NewLinker(parseTree *parser.ParseTree) *Linker {
	return &Linker{
		src:             parseTree,
		errors:          newLinkerErrorCollection(),
		dependencyGraph: make(dependencyGraph),
	}
}

func (self *Linker) LinkedScopes() scope.Scope {
	return self.rootScope
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
	self.verifyScopeObjects(self.rootScope.(*scope.PackageScope))
}

func (self *Linker) verifyScopeObjects(s *scope.PackageScope) {
	for _, child := range s.Children {
		if child.Kind() == scope.Package {
			self.verifyScopeObjects(child.(*scope.PackageScope))
		} else {
			m := map[string]*scope.Import{}

			// verify objects between imports
			for alias, imports := range (child.(*scope.FileScope)).Imports {
				if alias == "self" {
					continue
				}

				for _, imp := range *imports {
					for _, obj := range imp.ImportedScope.GetObjects(1) {
						if _, ok := m[obj.Name]; ok {
							// found another object and this import does not has an alias
							if alias == "." || len(alias) == 0 {
								self.errors.push(NewLinkerErr(ErrAlreadyDefined{obj.Name}, reference.NewReference(child.Path(), imp.Pos)))
								continue
							}
						}

						m[obj.Name] = &imp
					}
				}
			}

			// verify local objects against imports
			// objects are not verified against other in the same file scope because they are already verified at discover stage
			for _, obj := range child.GetObjects(1) {
				if imp, ok := m[obj.Name]; ok {

					// if the imported object does not have an alias, report error
					if !imp.HasAlias() {
						self.errors.push(NewLinkerErr(ErrAlreadyDefined{obj.Name}, reference.NewReference(child.Path(), imp.Pos)))
						continue
					}
				}
			}
		}
	}
}

func (self *Linker) verifyCircularDependencies() {
	self.buildDependencyGraph(self.rootScope)

	// Check for circular dependencies in each the file scope
	for _, fileScope := range self.rootScope.(*scope.PackageScope).Children {
		hasCircularDeps, conflictingImports := self.hasCircularDeps(fileScope)
		if hasCircularDeps {
			for _, conflict := range conflictingImports {
				self.errors.push(NewLinkerErr(ErrCircularDependency{
					Src:  &conflict.importer,
					Dest: &conflict.imported,
				}, reference.NewReference(conflict.importer.Path, reference.NewPos())))
			}
		}
	}
}

func (self *Linker) hasCircularDeps(s scope.Scope) (hasCircularDeps bool, conflictingImports []dependencyConflict) {
	visited := make(map[scope.Scope]bool)
	stack := make(map[scope.Scope]bool)
	conflictingImports = make([]dependencyConflict, 0)

	hasCircularDeps = self.checkCircularDeps(s, &visited, &stack, &conflictingImports)
	return
}

func (self *Linker) checkCircularDeps(s scope.Scope, visited *map[scope.Scope]bool, stack *map[scope.Scope]bool, conflictingImports *[]dependencyConflict) bool {
	(*visited)[s] = true
	(*stack)[s] = true

	// Traverse all the neighbors of the current scope
	for _, neighbor := range (self.dependencyGraph)[s] {
		// If the neighbor is not visited, recursively check for circular dependencies
		if !(*visited)[neighbor] && self.checkCircularDeps(neighbor, visited, stack, conflictingImports) {
			return true
		} else if (*stack)[neighbor] {
			// If the neighbor is already in the recursion stack, a cycle is detected
			conflict := dependencyConflict{
				importer: reference.File{Path: s.Path()},
				imported: reference.File{Path: neighbor.Path()},
			}
			if !self.hasConflictImport(conflictingImports, conflict) {
				*conflictingImports = append(*conflictingImports, conflict)
			}
			return true
		}
	}

	// remove the current scope from the stack
	(*stack)[s] = false
	return false
}

func (self *Linker) hasConflictImport(list *[]dependencyConflict, conflict dependencyConflict) bool {
	for _, item := range *list {
		if item.imported == conflict.imported && item.importer == conflict.importer {
			return true
		}
	}

	return false
}

func (self *Linker) buildDependencyGraph(s scope.Scope) {

	switch s := s.(type) {
	case *scope.PackageScope:
		for _, child := range s.Children {
			self.buildDependencyGraph(child)
			(self.dependencyGraph)[s] = append((self.dependencyGraph)[s], child)
		}

	case *scope.FileScope:
		for importsAlias, importedScopes := range s.Imports {
			if importsAlias == "self" {
				continue
			}

			for _, importedScope := range *importedScopes {
				(self.dependencyGraph)[s] = append((self.dependencyGraph)[s], importedScope.ImportedScope)
			}
		}
	}
}

func (self *Linker) resolveImports() {
	self.resolveImport(self.rootScope.(*scope.PackageScope))
}

func (self *Linker) resolveImport(packageScope *scope.PackageScope) {
	for _, childScope := range packageScope.Children {
		if childScope.Kind() == scope.Package {
			self.resolveImport(childScope.(*scope.PackageScope))
		} else {
			fileScope := childScope.(*scope.FileScope)
			fileScopeAst := fileScope.Ast()
			childScopePath := filepath.Dir(fileScope.Path())

			aliases := make(map[string]bool)
			for _, use := range fileScopeAst.UseStatements {
				path := use.Path.Token.Literal
				alias := "."

				// check if alias is not already in use
				if use.Alias != nil {
					alias = use.Alias.Token.Literal
					if _, ok := aliases[alias]; ok {
						self.errors.push(NewLinkerErr(ErrAliasAlreadyDefined{alias}, reference.NewReference(childScope.Path(), use.Alias.Pos)))
						continue
					}
					aliases[alias] = true
				}

				// check if impPath is not equal to pkgScope.Path
				// it would be a self import
				if path == childScopePath {
					self.errors.push(NewLinkerErr(ErrSelfImport{}, reference.NewReference(childScope.Path(), use.Path.Pos)))
					continue
				}

				// find scope
				importedScope := self.rootScope.FindByPath(path)
				if importedScope == nil {
					self.errors.push(NewLinkerErr(ErrPackageNotFound{path}, reference.NewReference(childScope.Path(), use.Path.Pos)))
					continue
				}

				fileScope.Imports.Push(alias, scope.NewImport(use.Path.Token.Literal, use.UnwrapAlias(), importedScope, use.Path.Pos))
			}

			// add its parent so it can access siblings types
			fileScope.Imports.Push("self", scope.NewImport(fileScope.Parent().Path(), "self", fileScope.Parent(), reference.NewPos()))
		}
	}
}

func (self *Linker) findScope(path string) scope.Scope {
	return self.rootScope.FindByPath(path)
}

func (self *Linker) buildScopes() {
	root := self.src.Root()
	self.rootScope = scope.NewPackageScope(root.Path, nil)

	root.Iter(func(pkgName string, node *parser.ParseNode) {
		self.createScope(pkgName, node, self.rootScope)
	})
}

func (self *Linker) createScope(packageName string, node *parser.ParseNode, parent scope.Scope) {

	packageScope := scope.NewPackageScope(node.Path, parent).(*scope.PackageScope)
	for _, ast := range node.AstList {
		fileScope := scope.NewFileScope(ast.File.Path, ast, packageScope).(*scope.FileScope)

		// push types
		for _, stmt := range ast.TypeStatements {
			obj := scope.NewObject(stmt)

			if _, ok := fileScope.Objects[obj.Name]; ok {
				src := obj.Source()
				self.errors.push(NewLinkerErr(ErrAlreadyDefined{obj.Name}, reference.NewReference(ast.File.Path, src.Name.Pos)))
				continue
			}

			fileScope.Objects[obj.Name] = obj
		}

		packageScope.Children = append(packageScope.Children, fileScope)
	}

	(parent.(*scope.PackageScope)).Children = append((parent.(*scope.PackageScope)).Children, packageScope)

	node.Iter(func(pkgName string, node *parser.ParseNode) {
		self.createScope(pkgName, node, packageScope)
	})
}
