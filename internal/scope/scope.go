package scope

import "tomasweigenast.com/nexema/tool/internal/parser"

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

func (self *LocalScope) FindObject(name, alias string) (obj *Object, needAlias bool) {
	candidates := make(match)

	// if alias is empty, lookup locally
	if len(alias) == 0 {
		localObj, ok := self.objects[name]
		if ok {
			candidates.push("", localObj)
		}
	}

	// lookup in imported types
	if len(self.resolvedScopes) > 0 {
		for resolvedScope, imp := range self.resolvedScopes {
			matches := resolvedScope.FindObjects(name)
			candidates.push(imp.Alias, matches...)
		}
	}

	count := candidates.count()
	if count == 0 {
		return nil, false
	} else if count == 1 {
		return candidates.single(alias), false
	} else {
		// decide
		if len(alias) == 0 {
			return nil, true
		}

		return candidates.single(alias), false
	}
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

func (self *Scope) FindObjects(name string) []*Object {
	arr := make([]*Object, 0)

	for _, ls := range self.localScopes {
		obj, ok := ls.objects[name]
		if ok {
			arr = append(arr, obj)
		}
	}

	return arr
}

// match is a map where each key represents an alias and the value is the list of objects under that alias
type match map[string][]*Object

func (self *match) push(alias string, obj ...*Object) {
	if _, ok := (*self)[alias]; !ok {
		(*self)[alias] = make([]*Object, 0)
	}

	(*self)[alias] = append((*self)[alias], obj...)
}

func (self *match) count() int {
	count := 0

	for _, arr := range *self {
		count += len(arr)
	}

	return count
}

func (self *match) single(alias string) *Object {
	objs, ok := (*self)[alias]
	if !ok {
		return nil
	}

	if len(objs) == 0 {
		return nil
	}

	return objs[0]
}
