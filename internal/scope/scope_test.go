package scope

import (
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/internal/parser"
)

func TestScope_FindObject(t *testing.T) {
	tests := []struct {
		name        string
		scope       Scope
		typeName    string
		alias       string
		wantObjects []*Object
	}{
		{
			name:     "single local match",
			typeName: "A",
			scope: &FileScope{
				Objects: map[string]*Object{
					"A": {Name: "A"},
				},
			},
			wantObjects: []*Object{{Name: "A"}},
		},
		{
			name:     "single local non match",
			typeName: "B",
			scope: &FileScope{
				Objects: map[string]*Object{
					"A": {Name: "A"},
				},
			},
			wantObjects: nil,
		},
		{
			name:     "single import match",
			typeName: "A",
			scope: &FileScope{
				Imports: importCollection{
					"": &[]Import{
						{
							Alias: "",
							ImportedScope: &FileScope{
								Objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					},
				},
			},
			wantObjects: []*Object{{Name: "A"}},
		},
		{
			name:     "single import with non matching alias",
			typeName: "A",
			scope: &FileScope{
				Imports: importCollection{
					"foo": &[]Import{
						{
							Alias: "foo",
							ImportedScope: &FileScope{
								Objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					},
				},
			},
			wantObjects: nil,
		},
		{
			name:     "single import with alias match",
			typeName: "A",
			scope: &FileScope{
				Imports: importCollection{
					"foo": &[]Import{
						{
							Alias: "foo",
							ImportedScope: &FileScope{
								Objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					},
				},
			},
			alias:       "foo",
			wantObjects: []*Object{{Name: "A"}},
		},
		{
			name:     "many imports without specifying alias",
			typeName: "A",
			scope: &FileScope{
				Imports: importCollection{
					"": &[]Import{
						{
							Alias: "",
							ImportedScope: &FileScope{
								Objects: map[string]*Object{
									"A": {Id: "2", Name: "A"},
								},
							},
						},
						{
							Alias: "",
							ImportedScope: &FileScope{
								Objects: map[string]*Object{
									"A": {Id: "1", Name: "A"},
								},
							},
						},
					},
				},
			},
			wantObjects: []*Object{
				{Name: "A", Id: "2"},
				{Name: "A", Id: "1"},
			},
		},
		{
			name:     "many imports specifying alias",
			typeName: "A",
			scope: &FileScope{
				Imports: importCollection{
					"first": &[]Import{
						{
							Alias: "first",
							ImportedScope: &FileScope{
								Objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					},
					"second": &[]Import{
						{
							Alias: "second",
							ImportedScope: &FileScope{
								Objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					},
				},
			},
			alias:       "second",
			wantObjects: []*Object{{Name: "A"}},
		},
		{
			name:     "imports with alias does not matter",
			typeName: "A",
			scope: &FileScope{
				Imports: importCollection{
					"first": &[]Import{
						{
							Alias: "first",
							ImportedScope: &FileScope{
								Objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					},
					"second": &[]Import{
						{
							Alias: "second",
							ImportedScope: &FileScope{
								Objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					},
				},
			},
			wantObjects: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			objs := test.scope.FindObject(test.typeName, test.alias)
			require.Equal(t, test.wantObjects, objs)
		})
	}
}

func TestScope_GetObjects(t *testing.T) {
	root := NewPackageScope("./", nil).(*PackageScope)
	user := NewPackageScope("./user", root).(*PackageScope)
	userChild := NewFileScope("./user/user.nex", &parser.Ast{}, user).(*FileScope)
	userChild.Objects["Baz"] = NewObject(parser.TypeStmt{})
	user.Children = append(user.Children, userChild)

	identity := NewPackageScope("./identity", root).(*PackageScope)
	bar := NewPackageScope("./identity/bar", identity).(*PackageScope)
	foo := NewFileScope("./identity/bar/foo.nex", &parser.Ast{}, bar).(*FileScope)
	foo.Objects["Baz"] = NewObject(parser.TypeStmt{})
	bar.Children = append(bar.Children, foo)
	identity.Children = append(identity.Children, bar)
	root.Children = append(root.Children, user, identity)

	objs := root.GetObjects(2)
	require.Len(t, objs, 1)

	objs = root.GetObjects(3)
	require.Len(t, objs, 2)

	objs = bar.GetObjects(1)
	require.Len(t, objs, 1)
}

// func TestScope_FindObject(t *testing.T) {
// 	root := NewPackageScope("./root", nil).(*PackageScope)

// 	identityScope := NewPackageScope("./root/identity", root).(*PackageScope)
// 	userScope := NewFileScope("./root/identity/user.nex", identityScope).(*FileScope)
// 	userScope.Objects["User"] = NewObject(parser.TypeStmt{})
// 	userScope.Objects["AccountDetails"] = NewObject(parser.TypeStmt{})
// 	userScope.Imports[identityScope.path] = identityScope

// 	accountTypeScope := NewFileScope("./root/identity/account_type.nex", identityScope).(*FileScope)
// 	accountTypeScope.Objects["AccountType"] = NewObject(parser.TypeStmt{})
// 	accountTypeScope.Objects["Permissions"] = NewObject(parser.TypeStmt{})

// 	identityScope.Children = append(identityScope.Children, userScope, accountTypeScope)

// 	commonScope := NewPackageScope("./root/common", root).(*PackageScope)
// 	entityScope := NewFileScope("./root/common/entity.nex", commonScope).(*FileScope)
// 	entityScope.Objects["Entity"] = NewObject(parser.TypeStmt{})

// 	commonScope.Children = append(commonScope.Children, entityScope)

// 	root.Children = append(root.Children, identityScope, commonScope)

// 	entityObj := identityScope.FindObject("AccountDetails")
// 	require.NotNil(t, entityObj)
// }
