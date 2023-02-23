package scope

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScope_FindObject(t *testing.T) {
	tests := []struct {
		name          string
		localScope    *LocalScope
		typeName      string
		alias         string
		wantObject    *Object
		wantNeedAlias bool
	}{
		{
			name:     "single local match",
			typeName: "A",
			localScope: &LocalScope{
				objects: map[string]*Object{
					"A": {Name: "A"},
				},
			},
			wantObject:    &Object{Name: "A"},
			wantNeedAlias: false,
		},
		{
			name:     "single local non match",
			typeName: "B",
			localScope: &LocalScope{
				objects: map[string]*Object{
					"A": {Name: "A"},
				},
			},
			wantObject:    nil,
			wantNeedAlias: false,
		},
		{
			name:     "single import match",
			typeName: "A",
			localScope: &LocalScope{
				resolvedScopes: map[*Scope]*Import{
					{
						localScopes: []*LocalScope{
							{
								objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					}: {},
				},
			},
			wantObject:    &Object{Name: "A"},
			wantNeedAlias: false,
		},
		{
			name:     "single import with non matching alias",
			typeName: "A",
			localScope: &LocalScope{
				resolvedScopes: map[*Scope]*Import{
					{
						localScopes: []*LocalScope{
							{
								objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					}: {
						Alias: "foo",
					},
				},
			},
			wantObject:    nil,
			wantNeedAlias: false,
		},
		{
			name:     "single import with alias match",
			typeName: "A",
			localScope: &LocalScope{
				resolvedScopes: map[*Scope]*Import{
					{
						localScopes: []*LocalScope{
							{
								objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					}: {
						Alias: "foo",
					},
				},
			},
			alias:         "foo",
			wantObject:    &Object{Name: "A"},
			wantNeedAlias: false,
		},
		{
			name:     "many imports without specifying alias",
			typeName: "A",
			localScope: &LocalScope{
				resolvedScopes: map[*Scope]*Import{
					{
						localScopes: []*LocalScope{
							{
								objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					}: {},
					{
						localScopes: []*LocalScope{
							{
								objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					}: {},
				},
			},
			wantObject:    nil,
			wantNeedAlias: true,
		},
		{
			name:     "many imports specifying alias",
			typeName: "A",
			localScope: &LocalScope{
				resolvedScopes: map[*Scope]*Import{
					{
						localScopes: []*LocalScope{
							{
								objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					}: {
						Alias: "first",
					},
					{
						localScopes: []*LocalScope{
							{
								objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					}: {
						Alias: "second",
					},
				},
			},
			alias:         "second",
			wantObject:    &Object{Name: "A"},
			wantNeedAlias: false,
		},
		{
			name:     "many imports need alias",
			typeName: "A",
			localScope: &LocalScope{
				resolvedScopes: map[*Scope]*Import{
					{
						localScopes: []*LocalScope{
							{
								objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					}: {
						Alias: "first",
					},
					{
						localScopes: []*LocalScope{
							{
								objects: map[string]*Object{
									"A": {Name: "A"},
								},
							},
						},
					}: {
						Alias: "second",
					},
				},
			},
			wantObject:    nil,
			wantNeedAlias: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			obj, needAlias := test.localScope.FindObject(test.typeName, test.alias)
			require.Equal(t, test.wantObject, obj)
			require.Equal(t, test.wantNeedAlias, needAlias)
		})
	}
}
