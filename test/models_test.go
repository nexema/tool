package test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"tomasweigenast.com/schema_interpreter/internal"
)

func TestUnmarshalMPackSchemaConfig(t *testing.T) {
	var tests = []struct {
		name      string
		input     string
		inputType string
		want      internal.MPackSchemaConfig
	}{
		{
			name:      "unmarshal json",
			inputType: "json",
			input:     `{"name":"my_amazing_project","author":"ImTheAuthor","version":"1.0.0","dependencies":["git:github.com/dep/myDependency","local:../dep/myDependency"],"skip":["skip/this/folder","skip/this/file","skip/all/**"],"generators":{"dart":{"out":"./dist/dart","options":["writeReflection"]},"csharp":{"out":"./dist/csharp","options":["omitReflection"]}}}`,
			want: internal.MPackSchemaConfig{
				Name:    "my_amazing_project",
				Author:  "ImTheAuthor",
				Version: "1.0.0",
				Dependencies: []internal.MPackSchemaConfigDependency{
					{Source: internal.Git, Path: "github.com/dep/myDependency"},
					{Source: internal.Local, Path: "../dep/myDependency"},
				},
				Skip: []string{"skip/this/folder", "skip/this/file", "skip/all/**"},
				Generators: internal.MPackSchemaConfigGeneratorMap{
					"dart": internal.MPackSchemaConfigGenerator{
						Out:     "./dist/dart",
						Options: []string{"writeReflection"},
					},
					"csharp": internal.MPackSchemaConfigGenerator{
						Out:     "./dist/csharp",
						Options: []string{"omitReflection"},
					},
				},
			},
		},
		{
			name:      "unmarshal yaml",
			inputType: "yaml",
			input: `
name: my_amazing_project
author: ImTheAuthor
version: 1.0.0

dependencies:
  - git:github.com/dep/myDependency
  - local:../dep/myDependency

skip:
  - skip/this/folder
  - skip/this/file
  - skip/all/**

generators:
  dart:
    out: ./dist/dart
    options:
      - writeReflection
  csharp:
    out: ./dist/csharp
    options:
      - omitReflection
`,
			want: internal.MPackSchemaConfig{
				Name:    "my_amazing_project",
				Author:  "ImTheAuthor",
				Version: "1.0.0",
				Dependencies: []internal.MPackSchemaConfigDependency{
					{Source: internal.Git, Path: "github.com/dep/myDependency"},
					{Source: internal.Local, Path: "../dep/myDependency"},
				},
				Skip: []string{"skip/this/folder", "skip/this/file", "skip/all/**"},
				Generators: internal.MPackSchemaConfigGeneratorMap{
					"dart": internal.MPackSchemaConfigGenerator{
						Out:     "./dist/dart",
						Options: []string{"writeReflection"},
					},
					"csharp": internal.MPackSchemaConfigGenerator{
						Out:     "./dist/csharp",
						Options: []string{"omitReflection"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r internal.MPackSchemaConfig
			var err error
			if tt.inputType == "json" {
				err = json.Unmarshal([]byte(tt.input), &r)
			} else {
				err = yaml.Unmarshal([]byte(tt.input), &r)
			}

			require.Nil(t, err)

			if diff := cmp.Diff(tt.want, r); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}

}
