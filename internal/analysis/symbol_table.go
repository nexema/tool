package analysis

import (
	"fmt"

	"github.com/mitchellh/hashstructure/v2"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
)

type symbolTable map[sourceRef][]symbol

func (st symbolTable) push(ref sourceRef, newSymbol symbol) {
	symbols := st[ref]
	symbols = append(symbols, newSymbol)
	st[ref] = symbols
}
func (st *symbolTable) symbolsOf(ref sourceRef) *[]symbol {
	value := (*st)[ref]
	return &value
}

// sourceRef represents the path to the data that contains a set of symbols
type sourceRef struct {
	file reference.File
}

// a symbol represents a type or a service
type symbol interface {
	// name returns the name of the symbol
	name() string
}

// typeSymbol represents a type statement
type typeSymbol struct {
	source     *parser.TypeStatement // the source TypeStatement
	symbolName string                // the name of the symbol
	id         uint64                // the id, usually the hash of the source
}

func (s *typeSymbol) name() string {
	return s.symbolName
}

func newTypeSymbol(statement *parser.TypeStatement) *typeSymbol {
	id, err := hashstructure.Hash(statement, hashstructure.FormatV2, nil)
	if err != nil {
		panic(fmt.Errorf("could not generate hash for TypeStatement: %s", err))
	}
	return &typeSymbol{
		source:     statement,
		symbolName: statement.Name.TokenLiteral(),
		id:         id,
	}
}
