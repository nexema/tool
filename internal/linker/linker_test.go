package linker

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/scope"
	"tomasweigenast.com/nexema/tool/internal/token"
)

func TestLinker_Link(t *testing.T) {
	tests := []struct {
		name     string
		input    func() *parser.ParseTree
		wantErrs LinkerErrorCollection
		run      bool
	}{
		{
			name: "valid link",
			input: func() *parser.ParseTree {
				tree := parser.NewParseTree()
				tree.Insert("common", newAst("common/address.nex", []string{"Address", "Coordinates"}, []string{"identity/user"}))
				tree.Insert("identity/user", newAst("identity/user/user.nex", []string{"User", "AccountType"}, []string{}))
				return tree
			},
			wantErrs: nil,
		},
		{
			name: "valid link 2",
			input: func() *parser.ParseTree {
				tree := parser.NewParseTree()
				tree.Insert("common", newAst("common/entity.nex", []string{"Address", "Coordinates"}, []string{}))
				tree.Insert("identity", newAst("identity/user.nex", []string{"User", "AccountType"}, []string{"common"}))
				return tree
			},
			wantErrs: nil,
		},
		{
			name: "self import",
			input: func() *parser.ParseTree {
				tree := parser.NewParseTree()
				tree.Insert("common", newAst("common/address.nex", []string{"Address", "Coordinates"}, []string{"common"}))
				return tree
			},
			wantErrs: LinkerErrorCollection{
				NewLinkerErr(ErrSelfImport{}, *reference.NewReference("common/address.nex", reference.NewPos(0, 0))),
			},
		},
		{
			name: "circular dependency",
			run:  true,
			input: func() *parser.ParseTree {
				tree := parser.NewParseTree()
				tree.Insert("common", newAst("common/address.nex", []string{"Address", "Coordinates"}, []string{"identity/user"}))
				tree.Insert("identity/user", newAst("identity/user/user.nex", []string{"User", "AccountType"}, []string{"common"}))
				return tree
			},
			wantErrs: LinkerErrorCollection{
				NewLinkerErr(ErrCircularDependency{
					Src:  &reference.File{Path: "identity/user/user.nex"},
					Dest: &reference.File{Path: "common"},
				}, *reference.NewReference("identity/user/user.nex", reference.NewPos(0, 0))),
			},
		},
		{
			name: "duplicated object names in same local scope",
			input: func() *parser.ParseTree {
				tree := parser.NewParseTree()
				tree.Insert("common", newAst("common/address.nex", []string{"Address", "Address"}, []string{}))
				return tree
			},
			wantErrs: LinkerErrorCollection{
				NewLinkerErr(ErrAlreadyDefined{
					Name: "Address",
				}, *reference.NewReference("", reference.NewPos(0, 0))),
			},
		},
		{
			name: "duplicated object names against imports without alias",
			input: func() *parser.ParseTree {
				tree := parser.NewParseTree()
				tree.Insert("common", newAst("common/address.nex", []string{"Address"}, []string{"identity"}))
				tree.Insert("identity", newAst("identity/user.nex", []string{"Address", "AccountType"}, []string{}))
				return tree
			},
			wantErrs: LinkerErrorCollection{
				NewLinkerErr(ErrAlreadyDefined{
					Name: "Address",
				}, *reference.NewReference("", reference.NewPos(0, 0))),
			},
		},
		{
			name: "duplicated names are allowed if no import",
			input: func() *parser.ParseTree {
				tree := parser.NewParseTree()
				tree.Insert("common", newAst("common/address.nex", []string{"Address"}, []string{}))
				tree.Insert("identity", newAst("identity/user.nex", []string{"Address", "AccountType"}, []string{}))
				return tree
			},
			wantErrs: nil,
		},
		{
			name: "duplicated names are allowed with import alias",
			input: func() *parser.ParseTree {
				tree := parser.NewParseTree()
				tree.Insert("common", newAst("common/address.nex", []string{"Address"}, []string{"identity:my_alias"}))
				tree.Insert("identity", newAst("identity/user.nex", []string{"Address", "AccountType"}, []string{}))
				return tree
			},
			wantErrs: nil,
		},
		{
			name: "duplicated names are not allowed between imported packages",
			input: func() *parser.ParseTree {
				tree := parser.NewParseTree()
				tree.Insert("common", newAst("common/address.nex", []string{"Address"}, []string{}))
				tree.Insert("identity", newAst("identity/user.nex", []string{"Address"}, []string{}))
				tree.Insert("foo", newAst("foo/bar.nex", []string{}, []string{"common", "identity"}))
				return tree
			},
			wantErrs: LinkerErrorCollection{
				NewLinkerErr(ErrAlreadyDefined{
					Name: "Address",
				}, *reference.NewReference("", reference.NewPos(0, 0))),
			},
		},
		{
			name: "duplicated alias in imports are not allowed",
			input: func() *parser.ParseTree {
				tree := parser.NewParseTree()
				tree.Insert("common", newAst("common/address.nex", []string{"Address"}, []string{"identity:foo", "another:foo"}))
				tree.Insert("identity", newAst("identity/user.nex", []string{}, []string{}))
				tree.Insert("another", newAst("another/admin.nex", []string{}, []string{}))
				return tree
			},
			wantErrs: LinkerErrorCollection{
				NewLinkerErr(ErrAliasAlreadyDefined{
					Alias: "foo",
				}, *reference.NewReference("", reference.NewPos(0, 0))),
			},
		},
		{
			name: "package not found",
			input: func() *parser.ParseTree {
				tree := parser.NewParseTree()
				tree.Insert("common", newAst("common/address.nex", []string{"Address"}, []string{"identity"}))
				return tree
			},
			wantErrs: LinkerErrorCollection{
				NewLinkerErr(ErrPackageNotFound{
					Name: "identity",
				}, *reference.NewReference("", reference.NewPos(0, 0))),
			},
		},
	}

	for _, test := range tests {
		if !test.run {
			continue
		}

		t.Run(test.name, func(t *testing.T) {
			linker := NewLinker(test.input())
			linker.Link()

			if test.wantErrs == nil {
				require.Empty(t, *linker.errors)
			} else {
				require.Equal(t, test.wantErrs, *linker.errors)
			}
		})
	}
}

func TestLinker_BuildScopes(t *testing.T) {
	tree := parser.NewParseTree()
	tree.Insert("common", newAst("common/address.nex", []string{"Address", "Coordinates"}, []string{"identity/user"}))
	tree.Insert("identity/user", newAst("identity/user/user.nex", []string{"User", "AccountType"}, []string{}))
	linker := NewLinker(tree)
	linker.buildScopes()

	require.Equal(t, ".", linker.rootScope.Name())
	require.Equal(t, "./", linker.rootScope.Path())
	root := (linker.rootScope.(*scope.PackageScope))

	require.Len(t, root.Children, 2)
	require.Equal(t, "common", root.Children[0].Name())
	require.Equal(t, scope.Package, root.Children[0].Kind())
	require.Len(t, (root.Children[0].(*scope.PackageScope)).Children, 1)
	require.Equal(t, "address.nex", (root.Children[0].(*scope.PackageScope)).Children[0].Name())
	require.Equal(t, scope.File, (root.Children[0].(*scope.PackageScope)).Children[0].Kind())
	require.Len(t, (root.Children[0].(*scope.PackageScope)).Children[0].GetObjects(1), 2)

	require.Equal(t, "identity", root.Children[1].Name())
	require.Equal(t, scope.Package, root.Children[1].Kind())
	require.Len(t, root.Children[1].(*scope.PackageScope).Children, 1)
	require.Equal(t, scope.Package, root.Children[1].(*scope.PackageScope).Children[0].Kind())
	require.Equal(t, "user", root.Children[1].(*scope.PackageScope).Children[0].Name())
	require.Len(t, (root.Children[1].(*scope.PackageScope).Children[0].(*scope.PackageScope)).Children, 1)
	require.Equal(t, scope.File, (root.Children[1].(*scope.PackageScope).Children[0].(*scope.PackageScope)).Children[0].Kind())
	require.Equal(t, "user.nex", (root.Children[1].(*scope.PackageScope).Children[0].(*scope.PackageScope)).Children[0].Name())
}

func newAst(fileName string, typeNames []string, uses []string) *parser.Ast {
	useStmts := []parser.UseStmt{}
	types := []parser.TypeStmt{}

	for _, name := range typeNames {
		types = append(types, parser.TypeStmt{
			Name: parser.IdentStmt{Token: *token.NewToken(token.Ident, name)},
		})
	}

	for _, use := range uses {
		parts := strings.Split(use, ":")
		path := parts[0]
		var alias *parser.IdentStmt
		if len(parts) == 2 {
			alias = &parser.IdentStmt{
				Token: *token.NewToken(token.Ident, parts[1]),
			}
		}

		useStmts = append(useStmts, parser.UseStmt{
			Path: parser.LiteralStmt{
				Token: *token.NewToken(token.String, path),
			},
			Alias: alias,
		})
	}

	return &parser.Ast{
		File:           &reference.File{Path: fileName},
		UseStatements:  useStmts,
		TypeStatements: types,
	}
}
