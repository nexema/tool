package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"tomasweigenast.com/schema_interpreter/internal"
)

func main() {
	helpText := `MessagePack Schema Interpreter - build .mpack schema files
Author: {{range .Authors}}{{ . }}{{end}}
License: {{.Copyright}}
Version: {{.Version}}

Commands:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{if .VisibleFlags}}{{end}}`

	builder := internal.NewBuilder()

	app := &cli.App{
		Suggest:               true,
		Version:               "1.0.0",
		Authors:               []*cli.Author{{Name: "Tomás Weigenast"}},
		Copyright:             "GPL-3.0 license",
		Name:                  "mpack",
		CustomAppHelpTemplate: helpText,
		Commands: []*cli.Command{
			{
				Name:  "build",
				Usage: "Builds the project and creates an output",

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "type",
						Value: "json",
						Usage: "the output type of the build [yaml|json]",
					},
					&cli.StringFlag{
						Name:  "output",
						Value: "console",
						Usage: "where the output of the build will be written [console|<file_path>]",
					},
				},
				Action: func(ctx *cli.Context) error {
					folderPath := ctx.Args().Get(0)
					outputType := ctx.String("type")
					outputDestination := ctx.String("output")
					if len(outputType) == 0 {
						outputType = "json"
					}

					if len(folderPath) == 0 {
						return cli.Exit("expected folder path", -1)
					}

					return internal.ConsoleBuild(builder, outputType, outputDestination, folderPath)
				},
			},
			{
				Name:      "generate",
				Usage:     "Builds the project and generates code for a programming language",
				ArgsUsage: "[input-path] [output-path]",
				Action: func(ctx *cli.Context) error {
					folderPath := ctx.Args().Get(0)
					outputPath := ctx.Args().Get(1)
					outputType := ctx.String("type")
					if len(outputType) == 0 {
						outputType = "json"
					}

					if len(folderPath) == 0 {
						return cli.Exit("expected folder path", -1)
					}

					if len(outputPath) == 0 {
						return cli.Exit("expected output path", -1)
					}

					err := internal.ConsoleGenerate(builder, folderPath, outputPath, outputType)
					if err != nil {
						return cli.Exit(err, -1)
					}

					fmt.Println("files generated successfully")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
