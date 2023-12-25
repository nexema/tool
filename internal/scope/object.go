package scope

import (
	"fmt"

	"github.com/mitchellh/hashstructure/v2"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
)

// Object represents an Ast TypeStatement
//
// It generates an unique id for it, based on the hashcode of the type
type Object struct {
	src  parser.TypeStatement
	Id   string // calculated as the hashcode of src
	Name string // extracted from src Name identifier
}

// NewObject creates a new Object from the given TypeStmt
func NewObject(from parser.TypeStatement) *Object {
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

func (self *Object) Source() *parser.TypeStatement {
	return &self.src
}

// Import represents an `include` statement that may be resolved
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
