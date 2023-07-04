package integration_test

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/project"
	"tomasweigenast.com/nexema/tool/internal/token"
	"tomasweigenast.com/nexema/tool/internal/utils"
)

func TestProjectBuilder(t *testing.T) {
	tests := []struct {
		name      string
		before    func()
		after     func(t *testing.T, projectBuilder *project.ProjectBuilder)
		inputPath string
	}{
		{
			name:      "successful build",
			inputPath: "./test0",
			before: func() {
				writeNexYaml("./test0", project.ProjectConfig{
					Version: 1,
				})
				writeFile("./test0/common/entity.nex", `
				type Entity base {
					id string
					created_at timestamp
					modified_at timestamp?
					deleted_at timestamp?
				}
				`)
				writeFile("./test0/identity/user.nex", `
				use "common"

				type User extends Entity {
					first_name string
					last_name string
					email string
					phone_number string?
					tags list(string)
					preferences map(string,bool)
				}
				`)
			},
			after: func(t *testing.T, projectBuilder *project.ProjectBuilder) {
				err := projectBuilder.Discover()
				require.NoError(t, err)

				err = projectBuilder.Build()
				require.NoError(t, err)

				err = projectBuilder.BuildSnapshot()
				require.NoError(t, err)

				snapshot := projectBuilder.GetSnapshot()
				require.NotNil(t, snapshot)
				require.Equal(t, []definition.NexemaFile{
					{
						Id:          "17079104774682735149",
						PackageName: "common",
						Path:        "common/entity.nex",
						Types: []definition.TypeDefinition{
							{
								Id:       "7757690481152332",
								Name:     "Entity",
								Modifier: token.Base,
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
									{
										Index: 2,
										Name:  "modified_at",
										Type:  definition.PrimitiveValueType{Primitive: definition.Timestamp, Nullable: true},
									},
									{
										Index: 3,
										Name:  "deleted_at",
										Type:  definition.PrimitiveValueType{Primitive: definition.Timestamp, Nullable: true},
									},
								},
							},
						},
					},
					{
						Id:          "14275443586636360348",
						PackageName: "identity",
						Path:        "identity/user.nex",
						Types: []definition.TypeDefinition{
							{
								Id:       "6841242565540347458",
								Name:     "User",
								Modifier: token.Struct,
								BaseType: utils.StringPtr("7757690481152332"),
								Fields: []*definition.FieldDefinition{
									{
										Index: 0,
										Name:  "first_name",
										Type:  definition.PrimitiveValueType{Primitive: definition.String},
									},
									{
										Index: 1,
										Name:  "last_name",
										Type:  definition.PrimitiveValueType{Primitive: definition.String},
									},
									{
										Index: 2,
										Name:  "email",
										Type:  definition.PrimitiveValueType{Primitive: definition.String},
									},
									{
										Index: 3,
										Name:  "phone_number",
										Type:  definition.PrimitiveValueType{Primitive: definition.String, Nullable: true},
									},
									{
										Index: 4,
										Name:  "tags",
										Type: definition.PrimitiveValueType{Primitive: definition.List, Arguments: []definition.BaseValueType{
											definition.PrimitiveValueType{Primitive: definition.String},
										}},
									},
									{
										Index: 5,
										Name:  "preferences",
										Type: definition.PrimitiveValueType{Primitive: definition.Map, Arguments: []definition.BaseValueType{
											definition.PrimitiveValueType{Primitive: definition.String},
											definition.PrimitiveValueType{Primitive: definition.Boolean},
										}},
									},
								},
							},
						},
					},
				}, snapshot.Files)
			},
		},
	}

	for _, testcase := range tests {
		t.Cleanup(func() {
			// cleanup
			err := os.RemoveAll(testcase.inputPath)
			if err != nil {
				t.Errorf("could not cleanup test %q: %v", testcase.name, err)
			}
		})
		t.Run(testcase.name, func(t *testing.T) {
			os.Mkdir(testcase.inputPath, os.ModePerm)
			testcase.before()
			projectBuilder := project.NewProjectBuilder(testcase.inputPath)
			testcase.after(t, projectBuilder)
		})
	}
}

func writeNexYaml(p string, cfg project.ProjectConfig) {
	buf, err := yaml.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(path.Join(p, "nexema.yaml"), buf, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func writeFile(p, contents string) {
	os.MkdirAll(path.Dir(p), os.ModePerm)
	os.WriteFile(p, []byte(contents), os.ModePerm)
}
