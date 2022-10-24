package internal

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

func (f *CompileResult) CreateJsonOutputIndented() string {
	buf, err := json.MarshalIndent(f, "", "    ")
	if err != nil {
		panic(err)
	}

	return string(buf)
}

func (f *CompileResult) CreateJsonOutput() string {
	buf, err := json.Marshal(f)
	if err != nil {
		panic(err)
	}

	return string(buf)
}

func (f *CompileResult) CreateYamlOutput() string {
	buf, err := yaml.Marshal(f)
	if err != nil {
		panic(err)
	}

	return string(buf)
}
