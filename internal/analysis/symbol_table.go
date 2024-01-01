package analysis

import (
	"fmt"

	"github.com/mitchellh/hashstructure/v2"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
)

// unaliased represents the given alias to those types that does not declare an specific alias
const unaliasedSymbols = ""
const selfSymbols = "__self__"

// symbolEntry represents a map where the key is the name of the symbol
type symbolEntry map[string]symbol

// symbolAlias represents a map where the key is the alias under the symbols are declared
type symbolAlias map[string]symbolEntry

// symbolTable tracks the list of symbols declared in sourceRef
type symbolTable map[reference.File]symbolAlias

func (st *symbolTable) push(ref reference.File, alias string, newSymbol symbol) {
	aliasedSymbols := (*st)[ref]
	if aliasedSymbols == nil {
		aliasedSymbols = make(symbolAlias)
	}

	symbols := aliasedSymbols[alias]
	if symbols == nil {
		symbols = make(symbolEntry)
	}

	symbolName := newSymbol.name()
	if _, ok := symbols[symbolName]; ok {
		panic("duplicated symbol name")
	}

	symbols[symbolName] = newSymbol
	aliasedSymbols[alias] = symbols
	(*st)[ref] = aliasedSymbols
}

func (st *symbolTable) pushAll(ref reference.File, alias string, se symbolEntry) {
	aliasedSymbols := (*st)[ref]
	if aliasedSymbols == nil {
		aliasedSymbols = make(symbolAlias)
	}

	symbols := aliasedSymbols[alias]
	if symbols == nil {
		symbols = make(symbolEntry)
	}

	for symbolName, newSymbol := range se {
		if _, ok := symbols[symbolName]; ok {
			panic(fmt.Errorf("duplicated symbol name: %q in alias %q", symbolName, alias))
		}

		symbols[symbolName] = newSymbol
	}

	aliasedSymbols[alias] = symbols
	(*st)[ref] = aliasedSymbols
}

func (sa *symbolTable) of(file reference.File) *symbolAlias {
	val := (*sa)[file]
	return &val
}

func (sa *symbolAlias) lookup(typeName, alias string) symbol {
	symbols := (*sa)[alias]
	if symbols == nil {
		return nil
	}

	return symbols[typeName]
}

func (st *symbolTable) getSelfSymbols(file reference.File) symbolEntry {
	aliased := (*st)[file]
	return aliased[selfSymbols]
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
