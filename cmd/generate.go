package cmd

import (
	"errors"
	"fmt"
	"os"
	p "path"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"tomasweigenast.com/nexema/tool/builder"
	"tomasweigenast.com/nexema/tool/definition"
)

func generateCmd(path, snapshotFile string, generateFor []string) error {
	builder := builder.NewBuilder(path)
	err := builder.Discover()
	if err != nil {
		return err
	}

	var snapshot *definition.NexemaSnapshot
	if len(snapshotFile) > 0 {
		fileContents, err := os.ReadFile(snapshotFile)
		if err != nil {
			return fmt.Errorf("could not read snapshot file. %s", err)
		}

		err = jsoniter.Unmarshal(fileContents, &snapshot)
		if err != nil {
			return fmt.Errorf("invalid Nexema Snapshot file. %s", err)
		}
	} else {
		err = builder.Build()
		if err != nil {
			return err
		}

		snapshot = builder.Snapshot()
	}

	// parse generators
	generators, err := getGenerators(generateFor)
	if err != nil {
		return err
	}

	cfg := builder.Config()
	pluginRequestBuffer, err := jsoniter.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("could not encode nexema snapshot, error: %s", err)
	}

	wroteCount := 0
	for pluginName, outputPath := range generators {
		if _, ok := cfg.Generators[pluginName]; !ok {
			return fmt.Errorf("plugin %s not defined in nexema.yaml", pluginName)
		}

		plugin, err := cfg.Generators.GetPlugin(pluginName)
		if err != nil {
			return err
		}

		result, err := plugin.Run(pluginRequestBuffer, []string{fmt.Sprintf(`--output-path=%s`, outputPath)})
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

			filepath := p.Join(outputPath, snapshotFile.Path, file.Name)
			outputFolder := p.Dir(filepath)
			err := os.MkdirAll(outputFolder, os.ModePerm)
			if err != nil {
				return fmt.Errorf("could not create output folder %q for file %q, error: %s", outputFolder, snapshotFile.FileName, err)
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
