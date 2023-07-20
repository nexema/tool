package scope

import (
	"path/filepath"

	"tomasweigenast.com/nexema/tool/internal/parser"
)

type ScopeKind int

const (
	Package ScopeKind = iota
	File
)

type importCollection map[string]*[]Import

type Scope interface {
	Name() string
	Path() string
	Kind() ScopeKind
	Parent() Scope
	FindByPath(path string) Scope

	// FindObject returns an slice of the matching objects with name and alias.
	// This always should return an slice of one object, but if there are more, it can be used
	// to alert the user with an error.
	FindObject(name, alias string) []*Object

	// GetObjects returns an slice of objects defined in the scope.
	GetObjects(maxDepth int) []*Object
}

type PackageScope struct {
	name     string
	path     string
	parent   Scope
	Children []Scope
}

type FileScope struct {
	name    string
	path    string
	parent  Scope
	ast     *parser.Ast
	Objects map[string]*Object // list of objects, where the key is the name of the object
	Imports importCollection   // imported scopes, where the key is the alias (dot "." when no alias is specified) and the value the list of scopes under the same alias
}

func NewPackageScope(path string, parent Scope) Scope {
	if path == "" {
		path = "./"
	}
	return &PackageScope{
		name:     filepath.Base(path),
		path:     path,
		parent:   parent,
		Children: make([]Scope, 0),
	}
}

func NewFileScope(path string, ast *parser.Ast, parent Scope) Scope {
	return &FileScope{
		name:    filepath.Base(path),
		path:    path,
		parent:  parent,
		ast:     ast,
		Objects: make(map[string]*Object),
		Imports: make(importCollection),
	}
}

func (self *PackageScope) Name() string {
	return self.name
}

func (self *PackageScope) Path() string {
	return self.path
}

func (self *PackageScope) Kind() ScopeKind {
	return Package
}

func (self *PackageScope) Parent() Scope {
	return self.parent
}

func (self *PackageScope) FindByPath(path string) Scope {
	if self.path == path {
		return self
	}

	for _, child := range self.Children {
		scope := child.FindByPath(path)
		if scope != nil {
			return scope
		}
	}

	return nil
}

func (self *PackageScope) FindObject(name, alias string) []*Object {
	for _, child := range self.Children {
		obj := child.FindObject(name, alias)
		if obj != nil {
			return obj
		}
	}

	return nil
}

func (self *PackageScope) GetObjects(maxDepth int) []*Object {
	objects := make([]*Object, 0)
	self.getObjectsRecursive(maxDepth, &objects)
	return objects
}

func (self *PackageScope) getObjectsRecursive(maxDepth int, objects *[]*Object) {
	if maxDepth == 0 {
		return
	}

	for _, child := range self.Children {
		*objects = append(*objects, child.GetObjects(maxDepth-1)...)
	}
}

func (self *FileScope) Name() string {
	return self.name
}

func (self *FileScope) Path() string {
	return self.path
}

func (self *FileScope) Kind() ScopeKind {
	return File
}

func (self *FileScope) Parent() Scope {
	return self.parent
}

func (self *FileScope) Ast() *parser.Ast {
	return self.ast
}

func (self *FileScope) FindByPath(path string) Scope {
	if self.path == path {
		return self
	}

	return nil
}

func (self *FileScope) FindObject(name, alias string) []*Object {
	visited := make(map[*FileScope]bool)
	return self.searchObject(name, alias, &visited)
}

func (self *FileScope) searchObject(name, alias string, visited *map[*FileScope]bool) []*Object {
	if alias == "" {
		alias = "."
	}

	if (*visited)[self] {
		return nil
	}
	(*visited)[self] = true
	matches := make([]*Object, 0)
	hasAlias := len(alias) > 0 && alias != "."

	// if alias is empty, lookup locally
	if !hasAlias {
		if obj, ok := self.Objects[name]; ok {
			matches = append(matches, obj)
		}
	}

	// lookup in imported types
	for importsAlias, importedScopes := range self.Imports {
		if (hasAlias && importsAlias != alias) || (!hasAlias && len(importsAlias) > 0 && importsAlias != ".") {
			continue
		}

		for _, imp := range *importedScopes {
			if imp.ImportedScope.Kind() == File {
				obj, ok := imp.ImportedScope.(*FileScope).Objects[name]
				if ok {
					matches = append(matches, obj)
				}
			} else {
				objects := imp.ImportedScope.GetObjects(1)
				for _, obj := range objects {
					matches = append(matches, obj)
				}
			}

		}
	}

	switch len(matches) {
	case 0:
		return nil
	default:
		return matches
	}
}

func (self *FileScope) GetObjects(maxDepth int) []*Object {
	objects := make([]*Object, len(self.Objects))
	i := 0
	for _, obj := range self.Objects {
		objects[i] = obj
		i++
	}

	return objects
}

// Push adds a new scope to an alias key. If the entry does not exist, its created first, otherwise, it appends the scope
func (self *importCollection) Push(alias string, scopeImport Import) {
	scopes, ok := (*self)[alias]
	if !ok {
		scopes = new([]Import)
		(*self)[alias] = scopes
	}

	*scopes = append(*scopes, scopeImport)
}
