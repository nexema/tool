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
	err := builder.Build()

	require.NoError(t, err)

	snapshot := builder.Snapshot()
	want := &definition.NexemaSnapshot{
		Version:  1,
		Hashcode: 1128978876879954002,
		Files: []definition.NexemaFile{
			{
				FileName:    "sample.nex",
				PackageName: "foo",
				Path:        "foo",
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
								Type:  definition.PrimitiveValueType{Primitive: definition.String},
							},
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
