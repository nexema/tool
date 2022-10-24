package internal

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

func NewDeclarationTree(key string, value DeclarationNode, children []*DeclarationTree) *DeclarationTree {
	return &DeclarationTree{
		Key:      key,
		Value:    value,
		Children: &children,
	}
}

type DeclarationTree struct {
	Value    DeclarationNode
	Key      string
	Children *[]*DeclarationTree
}

type DeclarationNode interface {
	Path() string
	Name() string
	IsPackage() bool
}

type PackageDeclarationNode struct {
	PackageName  string `json:"packageName" yaml:"packageName"` // The name of the package
	RelativePath string `json:"path" yaml:"path"`               // The path to the package
}

type FileDeclarationNode struct {
	Id       string            `json:"id" yaml:"id"`           // A random id for the file
	FilePath string            `json:"path" yaml:"path"`       // The path of the file
	Types    []*TypeDefinition `json:"types" yaml:"types"`     // The types in the file
	FileName string            `json:"-" yaml:"-"`             // The name of the file
	Imports  []string          `json:"imports" yaml:"imports"` // The packages imported
}

func NewPackageDeclarationNode(name, path string) PackageDeclarationNode {
	return PackageDeclarationNode{
		PackageName:  name,
		RelativePath: path,
	}
}

func NewFileDeclarationNode(id, path, name string, imports []string, types []*TypeDefinition) FileDeclarationNode {
	return FileDeclarationNode{
		Id:       id,
		FileName: name,
		FilePath: path,
		Types:    types,
		Imports:  imports,
	}
}

func (p PackageDeclarationNode) IsPackage() bool {
	return true
}

func (p PackageDeclarationNode) Path() string {
	return p.RelativePath
}

func (p PackageDeclarationNode) Name() string {
	return p.PackageName
}

func (p FileDeclarationNode) IsPackage() bool {
	return false
}

func (p FileDeclarationNode) Path() string {
	return p.FilePath
}

func (p FileDeclarationNode) Name() string {
	return p.FileName
}

func (t DeclarationTree) Sprint() string {
	buf := bytes.NewBufferString("tree: ")
	t.print(buf, true, "")

	return buf.String()
}

func (t DeclarationTree) Fprint(w io.Writer) {
	t.print(w, true, "")
}

func (t *DeclarationTree) print(w io.Writer, root bool, padding string) {
	if t == nil {
		return
	}

	index := 0

	if root {
		fmt.Fprintf(w, "%s%s :package\n", padding+getTreePrintPadding(root, getTreePrintBoxType(index, len(*t.Children))), t.Key)
		root = false
	}

	for _, v := range *t.Children {
		var tag string
		if v.Value.IsPackage() {
			tag = "package"
		} else {
			tag = "file"
		}

		fmt.Fprintf(w, "%s%s :%s\n", padding+getTreePrintPadding(root, getTreePrintBoxType(index, len(*t.Children))), v.Key, tag)
		v.print(w, false, padding+getTreePrintPadding(root, getTreePrintBoxTypeExternal(index, len(*t.Children))))
		index++
	}
}

func (t *DeclarationTree) Lookup(key string) (tree *DeclarationTree, ok bool) {
	frags := strings.Split(key, "/")
	return t.lookup(frags, true)
}

func (t *DeclarationTree) lookup(frags []string, root bool) (tree *DeclarationTree, ok bool) {
	fragsLen := len(frags)

	if fragsLen == 0 {
		return t, true
	}

	if fragsLen == 1 && t.Key == frags[0] {
		return t, true
	} else if root {
		frags = frags[1:]
	}

	if t.Children == nil {
		return nil, false
	}

	for _, v := range *t.Children {
		if v.Key == frags[0] {
			return v.lookup(frags[1:], false)
		}
	}

	return nil, false
}

func (t *DeclarationTree) Add(node DeclarationNode) {
	frags := strings.Split(node.Path(), "/")
	t.add(frags, node)
}

func (t *DeclarationTree) add(frags []string, node DeclarationNode) {
	if len(frags) == 0 {
		return
	}

	nextTree, ok := t.lookup(frags[:len(frags)-1], true)
	if ok {
		if !nextTree.Value.IsPackage() {
			panic("cannot append to a file node")
		}

		newTree := NewDeclarationTree(frags[len(frags)-1], node, make([]*DeclarationTree, 0))
		*nextTree.Children = append(*nextTree.Children, newTree)
	}
}
