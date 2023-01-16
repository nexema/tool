package internal

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildScopes(t *testing.T) {
	aAst := &Ast{
		File: &File{Name: "A.nex", Pkg: "."},
		Types: &[]*TypeStmt{
			getTypeStmt("TypeA", Token_Struct),
			getTypeStmt("TypeB", Token_Struct, getFieldStmt("my_field", "TypeA")),
		},
	}

	// b resides in the same package as a, so it can safely import TypeA without any import stmt
	bAst := &Ast{
		File: &File{Name: "B.nex", Pkg: "."},
		Imports: &[]*ImportStmt{
			{
				Path:  &IdentifierStmt{Lit: "c"},
				Alias: &IdentifierStmt{Lit: "my_c"},
			},
		},
		Types: &[]*TypeStmt{
			getTypeStmt("TypeC", Token_Struct, getFieldStmt("my_field", "TypeA")),
			getTypeStmt("TypeD", Token_Enum),
			getTypeStmt("Sub", Token_Union, getFieldStmt("field", "TypeD")), // this will collapse if does not use an alias
			//getTypeStmt("Sub", Token_Union, getFieldStmt("field", "c.TypeD")), // this will throw an error if uncommented
		},
	}

	// "c" resides in other package, so in order to let "b" import TypeD, it must use an importStmt
	cAst := &Ast{
		File: &File{Name: "C.nex", Pkg: "c"},
		Types: &[]*TypeStmt{
			getTypeStmt("TypeD", Token_Struct),
		},
	}

	// declares the same type's name that "c"
	dAst := &Ast{
		File: &File{Name: "D.nex", Pkg: "foo"},
		Types: &[]*TypeStmt{
			getTypeStmt("TypeD", Token_Struct),
		},
	}

	scopeCollection, err := BuildScopes(&AstTree{
		packageName: ".",
		sources: []*Ast{
			aAst,
			bAst,
		},
		children: []*AstTree{
			{
				packageName: "c",
				sources:     []*Ast{cAst},
			},
			{
				packageName: "foo",
				sources:     []*Ast{dAst},
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, scopeCollection)

	require.NotNil(t, scopeCollection.LookupPackageScope("c"))
	require.Nil(t, scopeCollection.LookupPackageScope("c/f"))

	obj, err := scopeCollection.LookupPackageScope(".").LookupObjectFor(bAst, "TypeD", "")
	require.Nil(t, err)
	require.NotNil(t, obj)
}

func getTypeStmt(name string, modifier Token, fields ...*FieldStmt) *TypeStmt {
	return &TypeStmt{
		Name:     &IdentifierStmt{Lit: name},
		Modifier: modifier,
		Fields:   &fields,
	}
}

func getFieldStmt(name, kind string) *FieldStmt {
	kindParts := strings.Split(kind, ".")
	var alias string
	if len(kindParts) == 2 {
		alias = kindParts[0]
		kind = kindParts[1]
	} else {
		kind = kindParts[0]
	}

	return &FieldStmt{
		Name: &IdentifierStmt{Lit: name},
		ValueType: &ValueTypeStmt{
			Ident: &IdentifierStmt{Lit: kind, Alias: alias},
		},
	}
}
