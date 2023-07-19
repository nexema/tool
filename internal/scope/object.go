package scope

import (
	"fmt"

	"github.com/mitchellh/hashstructure/v2"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
)

// Object represents an Ast Type statement.
type Object struct {
	src  parser.TypeStmt
	Id   string // calculated as the hashcode of src
	Name string // extracted from src Name identifier
}

// NewObject creates a new Object from the given TypeStmt
func NewObject(from parser.TypeStmt) *Object {
	name := from.Name.Token.Literal
	hashcode, err := hashstructure.Hash(from, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	return &Object{
		src:  from,
		Name: name,
		Id:   fmt.Sprint(hashcode),
	}
}

func (self *Object) Source() *parser.TypeStmt {
	return &self.src
}

// Import represents an `use` statement.
type Import struct {
	Pos           reference.Pos
	Path          string
	Alias         string
	ImportedScope Scope
}

func NewImport(path, alias string, scope Scope, pos reference.Pos) Import {
	return Import{
		Pos:           pos,
		Path:          path,
		ImportedScope: scope,
		Alias:         alias,
	}
}

func (self *Import) HasAlias() bool {
	return self.Alias != "." && self.Alias != ""
}
