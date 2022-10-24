package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"tomasweigenast.com/schema_interpreter/internal"
)

func getDeclarationTree() *internal.DeclarationTree {
	tree := internal.NewDeclarationTree("root", internal.NewPackageDeclarationNode("root_package", "root"), make([]*internal.DeclarationTree, 0))
	tree.Add(internal.NewFileDeclarationNode("1", "a.mpack", "a.mpack", make([]string, 0), &internal.TypeDefinitionCollection{}))
	tree.Add(internal.NewFileDeclarationNode("2", "b.mpack", "b.mpack", make([]string, 0), &internal.TypeDefinitionCollection{}))
	tree.Add(internal.NewPackageDeclarationNode("nested", "nested"))
	tree.Add(internal.NewFileDeclarationNode("3", "nested/c.mpack", "c.mpack", make([]string, 0), &internal.TypeDefinitionCollection{}))

	return tree
}

func TestReverseLookup(t *testing.T) {
	tree := getDeclarationTree()
	t.Log(tree.Sprint())
	node := (*tree.Children)[0]

	var err error
	var f *internal.DeclarationTree

	f, err = node.ReverseLookup("b.mpack")
	assert.Nil(t, err)

	if err == nil {
		assert.Equal(t, "b.mpack", f.Key)
	}

	node = (*(*tree.Children)[2].Children)[0]
	f, err = node.ReverseLookup("../a.mpack")
	assert.Nil(t, err)
	if err == nil {
		assert.Equal(t, "a.mpack", f.Key)
	}

	node = (*(*tree.Children)[2].Children)[0]
	_, err = node.ReverseLookup("../../a.mpack")
	assert.Error(t, err)
}
