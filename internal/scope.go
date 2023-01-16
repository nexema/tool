package internal

import (
	"errors"
	"fmt"
)

var (
	ErrNeedAlias error = errors.New("need alias")
	ErrNoObj     error = errors.New("no obj")
)

// PackageScope represents the declaration scope of one or more .nex files, a package.
// For example, two .nex files under X folder, has the same PackageScope, that means, each one
// can import types from the other, but they do not share the same LocalScope, meaning that if file Y
// imports package Z, file W will not be allowed to use types from package Z.
type PackageScope struct {
	PackageName  string
	Objects      *ObjectCollection    // list of objects declared by current PackageScope
	Participants map[*Ast]*LocalScope // participants of the current PackageScope
}

// LocalScope is the scope limited to the current Ast
type LocalScope struct {
	owner        *Ast
	root         *PackageScope
	Dependencies map[string]*scopeDependency // Dependencies contains the list of Objects imported by the key (import path)
}

type scopeDependency struct {
	Alias   string
	Objects *ObjectCollection
}

// ScopeCollection is a list of PackageScope
type ScopeCollection struct {
	Scopes []*PackageScope
	Tree   *AstTree
}

// EvaluateType evaluates a TypeStmt and returns an Object representing it
func EvaluateType(pkgName string, stmt *TypeStmt) *Object {
	return &Object{
		ID:   hashString(fmt.Sprintf("%s-%s", pkgName, stmt.Name.Lit)),
		Stmt: stmt,
	}
}

// NewPackageScope creates a new LocalScope
func NewPackageScope(packageName string) *PackageScope {
	return &PackageScope{
		PackageName:  packageName,
		Objects:      NewObjectCollection(),
		Participants: make(map[*Ast]*LocalScope),
	}
}

// addParticipant adds a new Ast and its dependencies to the LocalScope
func (s *PackageScope) addParticipant(ast *Ast, dependencies map[*Ast]string) error {
	// add current objects first
	for _, typeStmt := range *ast.Types {
		ok := s.Objects.append(typeStmt.Name.Lit, EvaluateType(s.PackageName, typeStmt))
		if !ok {
			return scopeErr(ast.File.Name, "type %q declared twice", typeStmt.Name.Lit)
		}
	}

	// add dependencies
	deps := make(map[string]*scopeDependency)
	for dep, alias := range dependencies {
		objCollection := NewObjectCollection()
		for _, typeStmt := range *dep.Types {
			typeName := typeStmt.Name.Lit

			// check if its not declared in local scope
			if s.Objects.exists(typeName) && alias == "" {
				return scopeErr(ast.File.Name, "type %q is declared in the current file but in %q package too, use an alias for your import", typeName, dep.File.Pkg)
			}

			// try to add to dependency scope
			ok := objCollection.append(typeName, EvaluateType(dep.File.Pkg, typeStmt))
			if !ok {
				return scopeErr("type %q is declared in %q package and in other imported package, use an alias for your imports", typeName, dep.File.Pkg)
			}
		}

		deps[dep.File.Pkg] = &scopeDependency{
			Alias:   alias,
			Objects: objCollection,
		}
	}
	s.Participants[ast] = &LocalScope{Dependencies: deps, root: s, owner: ast}
	return nil
}

// BuildScopes takes an astTree and outputs a list of Scope
func BuildScopes(astTree *AstTree) (*ScopeCollection, error) {
	scopeCollection := &ScopeCollection{
		Scopes: nil,
		Tree:   astTree,
	}

	builtScopes, err := scopeCollection.buildScopes(astTree)
	if err != nil {
		return nil, err
	}

	scopeCollection.Scopes = builtScopes
	return scopeCollection, nil
}

func (s *ScopeCollection) buildScopes(astTree *AstTree) ([]*PackageScope, error) {
	scopes := make([]*PackageScope, 0)

	scope := NewPackageScope(astTree.packageName)
	for _, src := range astTree.sources {
		deps := make(map[*Ast]string, 0)
		if src.Imports != nil {
			for _, imprt := range *src.Imports {
				importPath := imprt.Path.Lit

				// package cannot be imported itself
				if importPath == astTree.packageName {
					return nil, fmt.Errorf("the package %s cannot import itself", importPath)
				}

				// get sources
				astSources, ok := s.Tree.Lookup(importPath)
				if !ok {
					return nil, fmt.Errorf("package %q not found", importPath)
				}

				// iterate and add them
				for _, ast := range astSources {
					alias := ""
					if imprt.Alias != nil {
						alias = imprt.Alias.Lit
					}

					deps[ast] = alias
				}
			}
		}
		err := scope.addParticipant(src, deps)
		if err != nil {
			return nil, err
		}
	}
	scopes = append(scopes, scope)

	for _, child := range astTree.children {
		builtScopes, err := s.buildScopes(child)
		if err != nil {
			return nil, err
		}
		scopes = append(scopes, builtScopes...)
	}

	return scopes, nil
}

// LookupPackageScope looks up a PackageScope in the current ScopeCollection, searching by its import path
func (s *ScopeCollection) LookupPackageScope(packagePath string) *PackageScope {
	for _, scope := range s.Scopes {
		if scope.PackageName == packagePath {
			return scope
		}
	}

	return nil
}

// LookupObjectFor searches for an Object named name in the current PackageScope and for the given ast (for local package search).
// It will search first in s.Objects, if not found, it will start searching in dependencies.
// If it finds more than one match and no alias is specified, ErrNeedAlias will be returned.
// If it cannot find any match, ErrNoObj will be returned.
func (s *PackageScope) LookupObjectFor(ast *Ast, name, alias string) (*Object, error) {
	// matches reported, being map key the alias it was declared, and the value is the list of objects
	matches := make(matches)
	hasAlias := len(alias) > 0

	// search in local package
	obj := s.Objects.find(name)
	if obj != nil && !hasAlias {
		matches.add("", obj)
	}

	// search in dependencies
	localScope := s.Participants[ast]
	if localScope != nil {
		for _, dependency := range localScope.Dependencies {
			if hasAlias {
				if dependency.Alias != alias {
					continue
				}
			}

			obj = dependency.Objects.find(name)
			if obj != nil {
				matches.add(dependency.Alias, obj)
			}
		}
	}

	count := matches.count()
	if count == 0 {
		return nil, ErrNoObj
	} else if count == 1 {
		return matches.first(alias), nil
	} else {
		if len(alias) == 0 {
			objs := (matches)[alias]
			if len(*objs) > 1 {
				return nil, ErrNeedAlias
			}

			return (*objs)[0], nil
		}

		// this will be never reach
		panic("unexpected behaviour")
	}
}

// LookupObject looks up an object in the current local scope.
func (s *LocalScope) LookupObject(name, alias string) (*Object, error) {
	return s.root.LookupObjectFor(s.owner, name, alias)
}

func scopeErr(fileName, text string, args ...any) error {
	return fmt.Errorf(`%s -> %s`, fileName, fmt.Sprintf(text, args...))
}

type matches map[string]*[]*Object

func (m *matches) add(alias string, objs ...*Object) {
	arr, ok := (*m)[alias]
	if !ok {
		arr = new([]*Object)
		(*m)[alias] = arr
	}

	*arr = append(*arr, objs...)
}

func (m *matches) count() int {
	count := 0
	for _, obj := range *m {
		count += len(*obj)
	}

	return count
}

func (m *matches) first(alias string) *Object {
	return (*(*m)[alias])[0]
}
