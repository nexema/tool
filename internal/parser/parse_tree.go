package parser

import (
	"path/filepath"
	"strings"

	"github.com/tidwall/btree"
)

// ParseTree represents a parsed Ast tree structure
type ParseTree struct {
	root *ParseNode
}

// ParseNode represents a list of Ast files under a specific package path, and may have children
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

// Insert inserts the new ast to the specified path
func (self *ParseNode) Insert(path string, ast *Ast) {
	parts := strings.Split(path, "/")
	self.insert(parts, ast)
}

func (self *ParseNode) insert(parts []string, ast *Ast) {
	currentKey := parts[0]
	currentNode, ok := self.Children.Get(currentKey)
	if !ok {
		currentNode = NewParseNode()
		currentNode.Path = parts[0]
	}

	if len(parts) == 1 {
		path := ast.File.Path
		currentNode.Path = filepath.Dir(path)
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
	rootNode := NewParseNode()
	return &ParseTree{rootNode}
}

// Insert inserts the new ast to the specified path
func (self *ParseTree) Insert(path string, ast *Ast) {
	self.root.Insert(path, ast)
}

// Lookup looks up a ParseNode in the specified path
func (self *ParseTree) Lookup(path string) *ParseNode {
	return self.root.lookup(strings.Split(path, "/"))
}

// Iter iterates over every children in the ParseNode, executing f on each child
func (self *ParseNode) Iter(f func(pkgName string, node *ParseNode)) {
	self.Children.Scan(func(key string, value *ParseNode) bool {
		f(key, value)
		return true
	})
}
