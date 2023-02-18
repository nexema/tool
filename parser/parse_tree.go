package parser

import (
	"strings"

	"github.com/tidwall/btree"
)

type ParseTree struct {
	root *ParseNode
}

type ParseNode struct {
	Path     string
	AstList  []*Ast
	Children *btree.Map[string, *ParseNode]
}

func NewParseNode() *ParseNode {
	return &ParseNode{
		AstList:  make([]*Ast, 0),
		Children: new(btree.Map[string, *ParseNode]),
	}
}

func (self *ParseNode) Insert(path string, ast *Ast) {
	parts := strings.Split(path, "/")
	self.insert(parts, ast)
}

func (self *ParseNode) insert(parts []string, ast *Ast) {
	currentKey := parts[0]
	currentNode, ok := self.Children.Get(currentKey)
	if !ok {
		currentNode = NewParseNode()
	}

	if len(parts) == 1 {
		path := ast.File.Path
		currentNode.Path = path
		currentNode.AstList = append(currentNode.AstList, ast)
	} else {
		currentNode.insert(parts[1:], ast)
	}
	self.Children.Set(currentKey, currentNode)
}

func (self *ParseNode) lookup(parts []string) *ParseNode {
	currentKey := parts[0]
	currentNode, ok := self.Children.Get(currentKey)
	if !ok {
		return nil
	}

	if len(parts) == 1 {
		return currentNode
	} else {
		return currentNode.lookup(parts[1:])
	}
}

func NewParseTree() *ParseTree {
	return &ParseTree{NewParseNode()}
}

func (self *ParseTree) Insert(path string, ast *Ast) {
	self.root.Insert(path, ast)
}

func (self *ParseTree) Lookup(path string) *ParseNode {
	return self.root.lookup(strings.Split(path, "/"))
}
