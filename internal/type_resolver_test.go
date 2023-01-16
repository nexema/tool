package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAstTreeLookup(t *testing.T) {
	astTree := AstTree{
		packageName: ".",
		sources:     []*Ast{},
		children: []*AstTree{
			{
				packageName: "A",
				sources:     []*Ast{},
				children:    []*AstTree{},
			},
			{
				packageName: "B",
				sources:     []*Ast{},
				children: []*AstTree{
					{
						packageName: "C",
						sources:     []*Ast{},
					},
				},
			},
		},
	}

	_, ok := astTree.Lookup("A")
	require.True(t, ok)

	_, ok = astTree.Lookup("A/D")
	require.False(t, ok)

	_, ok = astTree.Lookup("B/C")
	require.True(t, ok)

	_, ok = astTree.Lookup("B/C/D")
	require.False(t, ok)
}

func TestAstTreeAppend(t *testing.T) {
	astTree := AstTree{
		packageName: ".",
		sources:     []*Ast{},
		children:    []*AstTree{},
	}

	astTree.append(&Ast{
		File: &File{Pkg: "testdata"},
	}, nil)

	astTree.append(&Ast{
		File: &File{Pkg: "testdata"},
	}, nil)
	astTree.append(&Ast{
		File: &File{Pkg: "testdata/baz"},
	}, nil)

	astTree.append(&Ast{
		File: &File{Pkg: "testdata/root/another"},
	}, nil)
	astTree.append(&Ast{
		File: &File{Pkg: "testdata/root/secondaty"},
	}, nil)

	astTree.append(&Ast{
		File: &File{Pkg: "testdata/root"},
	}, nil)

	astTree.print("")
}
