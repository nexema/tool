package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
	"tomasweigenast.com/nexema/tool/internal/token"
	"tomasweigenast.com/nexema/tool/internal/utils"
)

func TestBuilderSnapshot(t *testing.T) {
	scope0 := scope.NewScope(".", "root")
	scope1 := scope.NewScope("identity", "identity")

	commonScope := scope.NewLocalScope(&parser.File{
		Path: "./common.nex",
	}, make(map[string]*scope.Import), map[string]*scope.Object{
		"Entity": scope.NewObject(*utils.NewTypeBuilder("Entity").
			Modifier(token.Base).
			Field(utils.
				NewFieldBuilder("id").
				BasicValueType("string", false).
				Result()).
			Field(utils.
				NewFieldBuilder("created_at").
				BasicValueType("timestamp", false).
				Result()).
			Result()),
	})
	scope0.PushLocalScope(commonScope)

	usersScope := scope.NewLocalScope(&parser.File{
		Path: "./identity/users.nex",
	}, map[string]*scope.Import{
		"common.nex": utils.NewImport(utils.NewUseStmt("common.nex", "")),
	}, map[string]*scope.Object{
		"User": scope.NewObject(*utils.
			NewTypeBuilder("user").
			Modifier(token.Struct).
			Base("Entity").
			Field(utils.
				NewFieldBuilder("name").
				BasicValueType("string", false).
				Result()).
			Field(utils.
				NewFieldBuilder("phone_number").
				BasicValueType("string", true).
				Result()).
			Result()),
	})
	usersScope.AddResolvedScope(scope0, utils.NewImport(utils.NewUseStmt("common.nex", "")))
	scope1.PushLocalScope(usersScope)

	schemaBuilder := NewSchemaBuilder([]*scope.Scope{scope0, scope1})
	snapshot := schemaBuilder.BuildSnapshot()
	require.Equal(t, definition.NexemaSnapshot{
		Version:  1,
		Hashcode: "5133706252517793543",
		Files: []definition.NexemaFile{
			{
				Id:          "10629871956429266386",
				PackageName: "common.nex",
				Path:        "./common.nex",
				Types: []definition.TypeDefinition{
					{
						Id:       "3946789668698579447",
						Name:     "Entity",
						Modifier: token.Base,
						BaseType: nil,
						Fields: []*definition.FieldDefinition{
							{
								Index: 0,
								Name:  "id",
								Type:  definition.PrimitiveValueType{Primitive: definition.String},
							},
							{
								Index: 1,
								Name:  "created_at",
								Type:  definition.PrimitiveValueType{Primitive: definition.Timestamp},
							},
						},
					},
				},
			},
			{
				Id:          "9662408754454631521",
				PackageName: "users.nex",
				Path:        "./identity/users.nex",
				Types: []definition.TypeDefinition{
					{
						Id:       "1269049382580609917",
						Name:     "user",
						Modifier: token.Struct,
						BaseType: utils.StringPtr("3946789668698579447"),
						Fields: []*definition.FieldDefinition{
							{
								Index: 0,
								Name:  "name",
								Type:  definition.PrimitiveValueType{Primitive: definition.String},
							},
							{
								Index: 1,
								Name:  "phone_number",
								Type:  definition.PrimitiveValueType{Primitive: definition.String, Nullable: true},
							},
						},
					},
				},
			},
		},
	}, *snapshot)
}
