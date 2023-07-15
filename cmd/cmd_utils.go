package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
	"tomasweigenast.com/nexema/tool/internal/project"
)

func readCfg(p ...string) (config *project.ProjectConfig, projectFile string, err error) {
	var workingDir string
	if len(p) != 1 {
		workingDir = p[0]
	} else {
		workingDir, err = os.Getwd()
		if err != nil {
			return
		}
	}

	// read config first
	config = &project.ProjectConfig{}

	projectFile = path.Join(workingDir, "nexema.yaml")
	fileContent, err := ioutil.ReadFile(projectFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", fmt.Errorf("not in a nexema project directory")
		}

		return nil, "", fmt.Errorf("could not read nexema.yaml, error: %s", err)
	}

	err = yaml.Unmarshal(fileContent, config)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read nexema.yaml, error: %s", err)
	}

	return config, projectFile, nil
}
