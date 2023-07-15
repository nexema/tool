package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v2"
	"tomasweigenast.com/nexema/tool/internal/nexema"
	"tomasweigenast.com/nexema/tool/internal/project"
)

func modInit(ctx *cli.Context) error {
	p := ctx.Args().First()
	if len(p) == 0 || p == "." {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		p = dir
	}

	overwrite := ctx.Bool("overwrite")

	err := os.MkdirAll(p, os.ModePerm)
	if err != nil {
		return err
	}

	config := project.ProjectConfig{
		Version:    1,
		Name:       path.Base(p),
		Generators: make(project.NexemaGenerators),
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

func modAddGenerator(ctx *cli.Context) error {
	pluginName := ctx.Args().First()
	if pluginName == "" {
		return fmt.Errorf("the name of the plugin is required")
	}

	binPath := ctx.String("bin-path")

	config, projectFile, err := readCfg()
	if err != nil {
		return err
	}

	if _, ok := config.Generators[pluginName]; ok {
		return fmt.Errorf("generator for plugin %q already defined", pluginName)
	}

	// check if well known plugin exists
	if len(binPath) == 0 {
		_, err := nexema.GetWellKnownPlugin(pluginName)
		if err != nil {
			return err
		}
	}

	config.Generators[pluginName] = project.NexemaGenerator{
		Options: make(map[string]any),
		BinPath: binPath,
	}

	// write
	buffer, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(projectFile, buffer, os.ModePerm)
	if err != nil {
		return err
	}

	logrus.Infof("Generator for plugin %q added", pluginName)

	return nil
}

func modRemoveGenerator(ctx *cli.Context) error {
	pluginName := ctx.Args().First()
	if pluginName == "" {
		return fmt.Errorf("the name of the plugin is required")
	}

	config, projectFile, err := readCfg()
	if err != nil {
		return err
	}

	_, ok := config.Generators[pluginName]
	if !ok {
		return fmt.Errorf("no generator for plugin %q is defined", pluginName)
	}

	delete(config.Generators, pluginName)

	// write
	buffer, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(projectFile, buffer, os.ModePerm)
	if err != nil {
		return err
	}

	logrus.Infof("Generator for plugin %q added", pluginName)

	return nil
}
