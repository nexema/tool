package scope

import (
	"fmt"

	"github.com/mitchellh/hashstructure/v2"
	"tomasweigenast.com/nexema/tool/internal/parser"
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
	src   *parser.UseStmt
	Path  string
	Alias string
}

func NewImport(stmt *parser.UseStmt) *Import {
	var alias string
	if stmt.Alias != nil {
		alias = stmt.Alias.Token.Literal
	}

	return &Import{
		src:   stmt,
		Path:  stmt.Path.Token.Literal,
		Alias: alias,
	}
}

func (self *Import) Source() *parser.UseStmt {
	return self.src
}

func (self *Import) HasAlias() bool {
	return len(self.Alias) > 0
}
