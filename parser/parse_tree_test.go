package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTree_Insert(t *testing.T) {
	parseTree := NewParseTree()
	parseTree.Insert("identity", newAst("identity", "user.nex"))
	parseTree.Insert("identity", newAst("identity", "account.nex"))
	parseTree.Insert("common", newAst("common", "common.nex"))
	parseTree.Insert("common/address", newAst("common/address", "address.nex"))

	node := parseTree.Lookup("common")
	require.NotNil(t, node)
	require.Len(t, node.AstList, 1)
	require.Equal(t, 1, node.Children.Len())

	node = parseTree.Lookup("common/address")
	require.NotNil(t, node)
	require.Len(t, node.AstList, 1)
	require.Equal(t, 0, node.Children.Len())
}

func newAst(path, fileName string) *Ast {
	return &Ast{
		File: &File{
			Path:     path,
			FileName: fileName,
		},
	}
}
