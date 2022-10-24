package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

var (
	ErrPathOutsideContext error = errors.New("path points outside of the working context")
	ErrFileNotFound       error = errors.New("file not found")
)

type DeclarationTree struct {
	Value    DeclarationNode     `json:"value" yaml:"value"`
	Key      string              `json:"name" yaml:"name"`
	Children *[]*DeclarationTree `json:"children" yaml:"children"`
}

type DeclarationNode interface {
	Path() string
	Name() string
	IsPackage() bool
	Parent() *DeclarationTree
}

type PackageDeclarationNode struct {
	PackageName  string `json:"packageName" yaml:"packageName"` // The name of the package
	RelativePath string `json:"path" yaml:"path"`               // The path to the package

	parent *DeclarationTree `json:"-" yaml:"-"` // The parent node
}

type FileDeclarationNode struct {
	Id       string                    `json:"id" yaml:"id"`             // A random id for the file
	FilePath string                    `json:"path" yaml:"path"`         // The path of the file
	Types    *TypeDefinitionCollection `json:"types" yaml:"types"`       // The types in the file
	FileName string                    `json:"fileName" yaml:"fileName"` // The name of the file
	Imports  []string                  `json:"imports" yaml:"imports"`   // The packages imported
	parent   *DeclarationTree          `json:"-" yaml:"-"`               // The parent node
}

func NewDeclarationTree(key string, value DeclarationNode, children []*DeclarationTree) *DeclarationTree {
	return &DeclarationTree{
		Key:      key,
		Value:    value,
		Children: &children,
	}
}

func NewPackageDeclarationNode(name, path string) *PackageDeclarationNode {
	return &PackageDeclarationNode{
		PackageName:  name,
		RelativePath: path,
	}
}

func NewFileDeclarationNode(id, path, name string, imports []string, types *TypeDefinitionCollection) *FileDeclarationNode {
	return &FileDeclarationNode{
		Id:       id,
		FileName: name,
		FilePath: path,
		Types:    types,
		Imports:  imports,
	}
}

func (p PackageDeclarationNode) Parent() *DeclarationTree {
	return p.parent
}

func (p FileDeclarationNode) Parent() *DeclarationTree {
	return p.parent
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
	buf := bytes.NewBufferString("")
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

// Lookup looks up a DeclarationTree for a given path, taking t as the root package
func (t *DeclarationTree) Lookup(key string) (tree *DeclarationTree, ok bool) {
	frags := strings.Split(key, "/")
	return t.lookup(frags, true)
}

// ReverseLookup looks up a DeclarationTree for a given path, taking t as the origin
// For example, ../another.mpack will go back in tree two times, because the origin node
// is the file where the import is declared, and does not contain any children, so it must reverse to the parent package node.
func (t *DeclarationTree) ReverseLookup(path string) (tree *DeclarationTree, err error) {
	path = filepath.ToSlash(path)

	frags := strings.Split(path, "/")
	tree, err = t.reverseLookup(frags)
	if errors.Is(err, ErrPathOutsideContext) {
		return nil, fmt.Errorf("path %s points outside of the current working context", path)
	}

	if errors.Is(err, ErrFileNotFound) {
		return nil, fmt.Errorf("file %s not found at %s", path, t.Key)
	}

	return tree, err
}

func (t *DeclarationTree) reverseLookup(frags []string) (tree *DeclarationTree, err error) {
	fragsLen := len(frags)
	if fragsLen == 0 {
		panic("frags must be greater than 0")
	}

	origin := t

	// if not a package, go back one level to the reach its parent package
	if !t.Value.IsPackage() {
		origin = origin.Value.Parent()
	}

	// go back in tree while ".." appears
	for frags[0] == ".." {
		origin = origin.Value.Parent()
		frags = frags[1:]
		fragsLen--
	}

	if origin == nil {
		return nil, ErrPathOutsideContext
	}

	// search in current "origin" children
	for _, v := range *origin.Children {
		if v.Value.Name() == frags[0] {
			if fragsLen == 1 {
				return v, nil
			}

			return v.reverseLookup(frags[1:])
		}
	}

	return nil, ErrFileNotFound
}

func (t *DeclarationTree) lookup(frags []string, root bool) (tree *DeclarationTree, ok bool) {
	fragsLen := len(frags)

	if fragsLen == 0 {
		return t, true
	}

	if fragsLen == 1 && t.Key == frags[0] {
		return t, true
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
	frags := strings.Split(filepath.ToSlash(node.Path()), "/")
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

		fn, ok := node.(*FileDeclarationNode)
		if ok {
			fn.parent = nextTree
		}

		pn, ok := node.(*PackageDeclarationNode)
		if ok {
			pn.parent = nextTree
		}

		newTree := NewDeclarationTree(frags[len(frags)-1], node, make([]*DeclarationTree, 0))
		*nextTree.Children = append(*nextTree.Children, newTree)
	}
}
