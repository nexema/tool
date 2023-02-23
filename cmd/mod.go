package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"
	"tomasweigenast.com/nexema/tool/nexema"
)

func modInit(p string, overwrite bool) error {
	err := os.MkdirAll(p, os.ModePerm)
	if err != nil {
		return err
	}

	config := nexema.NexemaConfig{
		Version:    1,
		Name:       path.Base(p),
		Generators: make(nexema.NexemaGenerators),
	}

	out, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	p = path.Join(p, "nexema.yaml")
	_, err = os.Stat(p)
	if !errors.Is(err, os.ErrNotExist) && !overwrite {
		return fmt.Errorf("%s already exists. It you want to overwrite, run command with the --overwrite flag", p)
	}

	err = os.WriteFile(p, out, os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Printf("Wrote %s\n", p)

	return nil
}
