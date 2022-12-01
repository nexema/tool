package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"tomasweigenast.com/schema_interpreter/internal"
)

func TestParseMPackSchemaDefinition(t *testing.T) {
	input := `
name: myProject
author: Tomas Wegenast
version: 1.0.0

dependencies:
    - git:github.com/dep/myDependency
    - local:../dep/myDependency
    - core

skip:
    - models/skip_this.mpack
    - models/skip_the_folder

generators:
  - dart:
      options:
        packageName: test_files
  - csharp:
      path: ../path/to/plugin
      options:
        lowerCase: true
`

	var model internal.MPackSchemaDefinition
	err := yaml.Unmarshal([]byte(input), &model)

	assert.Nil(t, err)

	if err == nil {
		assert.Equal(t, model.ProjectName, "myProject")
		assert.Equal(t, model.Author, "Tomas Wegenast")
		assert.Equal(t, model.Version, "1.0.0")

		assert.Equal(t, model.SkipFiles, []string{
			"models/skip_this.mpack",
			"models/skip_the_folder",
		})

		assert.Equal(t, model.Dependencies, []internal.MPackDependency{
			{
				Source: internal.DependencySource_Git,
				Path:   "github.com/dep/myDependency",
			},
			{
				Source: internal.DependencySource_Local,
				Path:   "../dep/myDependency",
			},
			{
				Source: internal.DependencySource_Local,
				Path:   "__core",
			},
		})

		assert.Equal(t, model.Generators, internal.GeneratorList{
			{
				Name: "dart",
				Options: &map[string]interface{}{
					"packageName": "test_files",
				},
			},
			{
				Name: "csharp",
				Path: internal.GetAsPointer("../path/to/plugin"),
				Options: &map[string]interface{}{
					"lowerCase": true,
				},
			},
		})
	}
}
