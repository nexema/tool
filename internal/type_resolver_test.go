package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	rootAst := &Ast{
		file: &File{name: "root.nex", pkg: "."},
		imports: &[]*ImportStmt{
			{path: &IdentifierStmt{lit: "A"}},
			{path: &IdentifierStmt{lit: "B"}},
		},
	}
	aAst := &Ast{
		file: &File{name: "a.nex", pkg: "A"},
		imports: &[]*ImportStmt{
			{path: &IdentifierStmt{lit: "B"}, alias: &IdentifierStmt{lit: "my_b"}},
		},
	}
	bAst := &Ast{
		file: &File{name: "b.nex", pkg: "B"},
		imports: &[]*ImportStmt{
			{path: &IdentifierStmt{lit: "B/C"}},
		},
	}
	cAst := &Ast{
		file: &File{name: "c.nex", pkg: "C"},
		imports: &[]*ImportStmt{
			{path: &IdentifierStmt{lit: "B"}},
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
	require.Equal(t, "root.nex", rootCtx.owner.file.name)
	require.Equal(t, ".", rootCtx.owner.file.pkg)
	require.Len(t, rootCtx.imported, 2)
	require.Contains(t, rootCtx.imported, aAst)
	require.Contains(t, rootCtx.imported, bAst)

	aCtx := typeResolver.contexts[1]
	require.Equal(t, "a.nex", aCtx.owner.file.name)
	require.Equal(t, "A", aCtx.owner.file.pkg)
	require.Len(t, aCtx.imported, 1)
	require.Contains(t, aCtx.imported, bAst)
	require.Equal(t, "my_b", *aCtx.imported[bAst])

	bCtx := typeResolver.contexts[2]
	require.Equal(t, "b.nex", bCtx.owner.file.name)
	require.Equal(t, "B", bCtx.owner.file.pkg)
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
