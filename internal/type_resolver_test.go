package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	rootAst := &Ast{
		File: &File{Name: "root.nex", Pkg: "."},
		Imports: &[]*ImportStmt{
			{Path: &IdentifierStmt{Lit: "A"}},
			{Path: &IdentifierStmt{Lit: "B"}},
		},
	}
	aAst := &Ast{
		File: &File{Name: "a.nex", Pkg: "A"},
		Imports: &[]*ImportStmt{
			{Path: &IdentifierStmt{Lit: "B"}, Alias: &IdentifierStmt{Lit: "my_b"}},
		},
	}
	bAst := &Ast{
		File: &File{Name: "b.nex", Pkg: "B"},
		Imports: &[]*ImportStmt{
			{Path: &IdentifierStmt{Lit: "B/C"}},
		},
	}
	cAst := &Ast{
		File: &File{Name: "c.nex", Pkg: "C"},
		Imports: &[]*ImportStmt{
			{Path: &IdentifierStmt{Lit: "B"}},
		},
	}

	astTree := &AstTree{
		packageName: ".",
		sources:     []*Ast{rootAst},
		children: []*AstTree{
			{
				packageName: "A",
				sources:     []*Ast{aAst},
				children:    []*AstTree{},
			},
			{
				packageName: "B",
				sources:     []*Ast{bAst},
				children: []*AstTree{
					{
						packageName: "C",
						sources:     []*Ast{cAst},
					},
				},
			},
		},
	}

	typeResolver := NewTypeResolver(astTree)
	typeResolver.Resolve()

	rootCtx := typeResolver.contexts[0]
	require.Equal(t, "root.nex", rootCtx.owner.File.Name)
	require.Equal(t, ".", rootCtx.owner.File.Pkg)
	require.Len(t, rootCtx.imported, 2)
	require.Contains(t, rootCtx.imported, aAst)
	require.Contains(t, rootCtx.imported, bAst)

	aCtx := typeResolver.contexts[1]
	require.Equal(t, "a.nex", aCtx.owner.File.Name)
	require.Equal(t, "A", aCtx.owner.File.Pkg)
	require.Len(t, aCtx.imported, 1)
	require.Contains(t, aCtx.imported, bAst)
	require.Equal(t, "my_b", *aCtx.imported[bAst])

	bCtx := typeResolver.contexts[2]
	require.Equal(t, "b.nex", bCtx.owner.File.Name)
	require.Equal(t, "B", bCtx.owner.File.Pkg)
	require.Len(t, bCtx.imported, 1)
	require.Contains(t, bCtx.imported, cAst)

	require.Len(t, typeResolver.errors, 1)
	require.Equal(t, "circular dependency between B and C not allowed", typeResolver.errors[0].Error())
}

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
