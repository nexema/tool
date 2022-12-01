package internal

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func ConsoleGenerate(builder *Builder, folderPath string, outputPath string, outputType string) error {
	output, err := builder.Build(folderPath)
	if err != nil {
		return cli.Exit(err, -1)
	}

	output.OutputPath = outputPath
	for _, generator := range builder.mpack.Generators {
		if generator.Name == "dart" {
			output.GeneratorOptions = generator.Options
		}
	}

	var outputString string
	switch outputType {
	case "json":
		outputString = output.CreateJsonOutput()

	case "yaml":
		outputString = output.CreateYamlOutput()

	default:
		return fmt.Errorf("unknown output type: %s", outputType)
	}

	// D:\Git\messagepack-schema\plugins\csharp\MessagePackSchema.Generator\bin\Debug\net6.0
	// plugin := NewPlugin("dotnet", "../plugins/csharp/MessagePackSchema.Generator/bin/Debug/net6.0/MessagePackSchema.Generator.dll")
	plugin := NewPlugin("mpack_generator")
	serialized := EncodeBase64(outputString)

	return plugin.Call(outputType, serialized)
}

func ConsoleBuild(builder *Builder, outputType string, outputDestination string, folderPath string) error {
	output, err := builder.Build(folderPath)
	if err != nil {
		return cli.Exit(err, -1)
	}
	var outputString string
	switch outputType {
	case "json":
		outputString = output.CreateJsonOutput()

	case "yaml":
		outputString = output.CreateYamlOutput()

	default:
		return fmt.Errorf("unknown output type: %s", outputType)
	}

	// get output destination
	switch outputDestination {
	case "console":
		fmt.Println(outputString)
		fmt.Printf("BASE 64: %v\n", EncodeBase64(outputString))

	default:
		// its a file
		err := os.WriteFile(outputDestination, []byte(outputString), 0644)
		if err != nil {
			return fmt.Errorf("cannot write output to file %s: %s", outputDestination, err)
		}
	}

	return nil
}
