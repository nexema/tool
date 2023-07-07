package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/project"
)

func build(ctx *cli.Context) error {
	inputPath := ctx.Args().First()
	if len(inputPath) == 0 {
		var err error
		inputPath, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	outputPath := ctx.String("out") // for snapshot if needed

	builder := project.NewProjectBuilder(inputPath)
	err := builder.Discover()
	if err != nil {
		return err
	}

	logrus.Infoln("Building project...")
	err = builder.Build()
	if err != nil {
		if errors.Is(err, project.ErrEmptyParseTree) {
			logrus.Infoln("Nothing to build")
			return nil
		}
		return err
	}

	// create snapshot
	logrus.Infof("Creating snapshot...")
	err = builder.BuildSnapshot()
	if err != nil {
		return err
	}

	if len(outputPath) > 0 {
		logrus.Infoln("Saving snapshot...")
		filepath, err := builder.SaveSnapshot(outputPath)
		if err != nil {
			return err
		}

		logrus.Infof("Snapshot file saved at %s\n", filepath)

		return nil
	}

	logrus.Infoln("Project built succcessfully.")

	return nil
}

func generate(ctx *cli.Context) error {
	snapshotPath := ctx.String("snapshot-file")
	inputPath := ctx.Args().First()
	if len(inputPath) == 0 {
		var err error
		inputPath, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	generateFor := ctx.StringSlice("for")
	projectBuilder := project.NewProjectBuilder(inputPath)
	err := projectBuilder.Discover()
	if err != nil {
		return err
	}

	var snapshot *definition.NexemaSnapshot
	if len(snapshotPath) > 0 {
		logrus.Infof("Generating from snapshot %q \n", snapshotPath)

		fileContents, err := os.ReadFile(snapshotPath)
		if err != nil {
			return fmt.Errorf("could not read snapshot file. %s", err)
		}

		err = jsoniter.Unmarshal(fileContents, &snapshot)
		if err != nil {
			return fmt.Errorf("invalid Nexema Snapshot file. %s", err)
		}
	} else {
		logrus.Infoln("Creating a fresh build...", snapshotPath)
		err = projectBuilder.Build()
		if err != nil {
			if errors.Is(err, project.ErrEmptyParseTree) {
				logrus.Infoln("Not in a Nexema project directory")
				return nil
			}
			return err
		}

		logrus.Infoln("Creating snapshot...")
		err = projectBuilder.BuildSnapshot()
		if err != nil {
			return err
		}

		snapshot = projectBuilder.GetSnapshot()
	}

	// parse generators
	generators, err := getGenerators(generateFor)
	if err != nil {
		return err
	}

	cfg := projectBuilder.GetConfig()
	pluginRequestBuffer, err := jsoniter.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("could not encode nexema snapshot, error: %s", err)
	}

	wroteCount := 0
	for pluginName, outputPath := range generators {
		logrus.Infof("Generating for %s...\n", pluginName)

		if _, ok := cfg.Generators[pluginName]; !ok {
			return fmt.Errorf("plugin %s not defined in nexema.yaml", pluginName)
		}

		plugin, err := cfg.Generators.GetPlugin(pluginName)
		if err != nil {
			return err
		}

		result, err := plugin.Run(pluginRequestBuffer, []string{fmt.Sprintf(`--output-path=%s`, outputPath)}, nil)

		if err != nil {
			return err
		}

		if result.ExitCode != 0 {
			var err string
			if result.Error != nil {
				err = *result.Error
			} else {
				err = "no error specified"
			}
			return fmt.Errorf("plugin %q failed with exit code %d (%s)", pluginName, result.ExitCode, err)
		}

		if result.Files == nil {
			return fmt.Errorf("no error was returned from the plugin but no file was returned too")
		}

		// start writing each file to its location
		for _, file := range *result.Files {
			// get file in snapshot
			snapshotFile := snapshot.FindFile(file.Id)
			if snapshotFile == nil {
				return fmt.Errorf("unable to find file %q", file.Id)
			}

			filepath := path.Join(outputPath, snapshotFile.Path, file.Name)
			outputFolder := path.Dir(filepath)
			err := os.MkdirAll(outputFolder, os.ModePerm)
			if err != nil {
				return fmt.Errorf("could not create output folder %q for file %q, error: %s", outputFolder, snapshotFile.Path, err)
			}

			err = os.WriteFile(filepath, []byte(file.Contents), os.ModePerm)
			if err != nil {
				return fmt.Errorf("could not write file %q, error: %s", filepath, err)
			}
			wroteCount++
		}
	}

	fmt.Printf("Wrote %d files successfully.\n", wroteCount)

	return nil
}

func format(ctx *cli.Context) error {
	path := ctx.String("path")
	if path == "" {
		return cli.Exit("path is required", 1)
	}
	fmt.Printf("Formatting code for the project at %s...\n", path)
	return nil
}

func getGenerators(in []string) (map[string]string, error) {
	out := make(map[string]string, len(in))
	for _, s := range in {
		toks := strings.Split(s, "=")
		if len(toks) != 2 {
			return nil, errors.New("invalid for argument, expected --for=[plugin-name]=[output-folder]")
		}

		out[toks[0]] = toks[1]
	}

	return out, nil
}
