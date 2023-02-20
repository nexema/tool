package scope

import "tomasweigenast.com/nexema/tool/parser"

// Scope represents a collection of local scopes, a.k.a .nex files.
// So, Scope is a Nexema package.
// Scope forgets any folder structure until it was created. That is done
// because any LocalScope can import any Scope
type Scope struct {
	path        string
	name        string
	localScopes []*LocalScope
}

// LocalScope represents a Nexema file, which contains a list of objects
// an may import other Scopes
type LocalScope struct {
	file           *parser.File
	imports        map[string]*Import
	objects        map[string]*Object
	resolvedScopes map[*Scope]*Import
}

func NewLocalScope(file *parser.File, imports map[string]*Import, objects map[string]*Object) *LocalScope {
	return &LocalScope{
		file:           file,
		imports:        imports,
		objects:        objects,
		resolvedScopes: make(map[*Scope]*Import),
	}
}

func (self *LocalScope) Objects() *map[string]*Object {
	return &self.objects
}

func (self *LocalScope) Imports() *map[string]*Import {
	return &self.imports
}

func (self *LocalScope) AddResolvedScope(scope *Scope, imp *Import) {
	self.resolvedScopes[scope] = imp
}

func (self *LocalScope) ResolvedScopes() *map[*Scope]*Import {
	return &self.resolvedScopes
}

func (self *LocalScope) File() *parser.File {
	return self.file
}

func NewScope(path, packageName string) *Scope {
	return &Scope{
		path:        path,
		name:        packageName,
		localScopes: make([]*LocalScope, 0),
	}
}

func (self *Scope) PushLocalScope(localScope *LocalScope) {
	self.localScopes = append(self.localScopes, localScope)
}

func (self *Scope) LocalScopes() *[]*LocalScope {
	return &self.localScopes
}

func (self *Scope) Path() string {
	return self.path
}

func (self *Scope) GetAllObjects() []*Object {
	arr := make([]*Object, 0)
	for _, local := range self.localScopes {
		for _, obj := range local.objects {
			arr = append(arr, obj)
		}
	}

	return arr
}
