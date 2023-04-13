package integration_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/builder"
	"tomasweigenast.com/nexema/tool/definition"
	"tomasweigenast.com/nexema/tool/token"
)

func TestBuilder_Build(t *testing.T) {
	builder := builder.NewBuilder("sample-project")
	err := builder.Discover()
	require.NoError(t, err)

	err = builder.Build()

	require.NoError(t, err)

	snapshot := builder.Snapshot()
	want := &definition.NexemaSnapshot{
		Version:  1,
		Hashcode: "16740850243460541223",
		Files: []definition.NexemaFile{
			{
				FileName:    "sample.nex",
				PackageName: "foo",
				Path:        "foo",
				Id:          "2472758977982914045",
				Types: []definition.TypeDefinition{
					{
						Name:     "Sample",
						Modifier: token.Struct,
						Fields: []*definition.FieldDefinition{
							{
								Name:  "id",
								Index: 0,
								Type:  definition.PrimitiveValueType{Primitive: definition.String},
							},
							{
								Name:  "name",
								Index: 1,
								Type:  definition.CustomValueType{ObjectId: "15292658956539431885"},
							},
							{
								Name:  "a_map",
								Index: 2,
								Type: definition.PrimitiveValueType{
									Primitive: "map",
									Arguments: []definition.BaseValueType{
										definition.PrimitiveValueType{Primitive: definition.String},
										definition.CustomValueType{ObjectId: "15292658956539431885"},
									},
								},
							},
						},
					},
					{
						Name:     "Another",
						Modifier: token.Enum,
						Fields: []*definition.FieldDefinition{
							{Name: "unknown"},
							{Name: "red", Index: 1},
							{Name: "blue", Index: 2},
						},
					},
				},
			},
		},
	}
	require.NotNil(t, snapshot)
	if diff := cmp.Diff(want, snapshot); len(diff) > 0 {
		t.Errorf("TestBuilder_Build: mismatch %s", diff)
	}
}
