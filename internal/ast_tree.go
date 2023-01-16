package internal

import (
	"fmt"
	"strings"
)

type AstTree struct {
	packageName string
	sources     []*Ast
	children    []*AstTree
}

// NewAstTree builds an AstTree from a list of Ast
func NewAstTree(astList []*Ast) *AstTree {
	root := &AstTree{
		sources:     make([]*Ast, 0),
		children:    make([]*AstTree, 0),
		packageName: ".",
	}
	for _, ast := range astList {
		file := ast.File
		if file.Pkg == "." { // root folder
			root.packageName = "."
			root.sources = append(root.sources, ast)
		} else {
			// nested tree
			root.append(ast, strings.Split(ast.File.Pkg, "/"))
		}
	}

	return root
}

// append adds a new Ast to s
func (s *AstTree) append(ast *Ast, frags []string) {
	if len(frags) == 0 {
		frags = strings.Split(ast.File.Pkg, "/")
		// frags = frags[len(frags)-1:]
	}

	if len(frags) == 1 {
		for _, child := range s.children {
			if child.packageName == frags[0] {
				child.sources = append(child.sources, ast)
				return
			}
		}

		// if not found, add new children
		s.children = append(s.children, &AstTree{
			packageName: frags[0],
			sources:     []*Ast{ast},
			children:    make([]*AstTree, 0),
		})
		return
	}

	path := frags[0]
	for _, child := range s.children {
		if child.packageName == path {
			child.append(ast, frags[1:])
			return
		}
	}

	// not found, add new children
	node := &AstTree{
		packageName: path,
		sources:     make([]*Ast, 0),
		children:    make([]*AstTree, 0),
	}
	node.append(ast, frags[1:])
	s.children = append(s.children, node)
}

// print prints the tree in a readable way
func (s *AstTree) print(tab string) {
	fmt.Printf("%s- %s\n", tab, s.packageName)
	for _, child := range s.children {
		child.print(tab + " ")
	}
}

// Lookup iterates over the AstTree and returns the list of Ast that are in the given packageName
func (s *AstTree) Lookup(packageName string) ([]*Ast, bool) {
	folders := strings.Split(packageName, "/")
	return s.lookup(folders)
}

func (s *AstTree) lookup(frags []string) ([]*Ast, bool) {
	if len(frags) == 0 {
		return s.sources, true
	}

	if len(frags) == 1 && s.packageName == frags[0] {
		return s.sources, true
	}

	for _, child := range s.children {
		if child.packageName == frags[0] {
			return child.lookup(frags[1:])
		}
	}

	return nil, false
}
