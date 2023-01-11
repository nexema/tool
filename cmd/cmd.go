package cmd

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

const helpText = `Nexema Generation Tool - build .nex schema files

Author: {{range .Authors}}{{ . }}{{end}}
License: {{.Copyright}}
Version: {{.Version}}

Commands:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{if .VisibleFlags}}{{end}}`

var commands []*cli.Command

func init() {
	buildCmd := &cli.Command{
		Name:            "build",
		Usage:           "builds a project and outputs a snapshot",
		Description:     "builds a Nexema project but does not generate any source file. Instead, generates a snapshot for later use or storage.",
		UsageText:       "nexema build [input-folderpath] [output-folderpath]",
		HideHelpCommand: true,
		HideHelp:        false,
		Action:          buildCmdExecutor,
	}

	generateCmd := &cli.Command{
		Name:            "generate",
		Usage:           "builds a project and generates source code files",
		Description:     "builds a Nexema project and generates source code files for all the targets specified in nexema.yaml.",
		UsageText:       "nexema generate [input-folderpath] [output-folderpath]",
		HideHelpCommand: true,
		HideHelp:        false,
		Action:          generateCmdExecutor,
	}

	formatCmd := &cli.Command{
		Name:            "format",
		Usage:           "formats all nexema files in the current project",
		UsageText:       "nexema format [input-folderpath]",
		HideHelpCommand: true,
		HideHelp:        false,
		Action:          formatCmdExecutor,
	}

	commands = []*cli.Command{
		buildCmd,
		generateCmd,
		formatCmd,
	}
}

func Run() {

	app := &cli.App{
		Suggest:               true,
		Version:               "1.0.0",
		Authors:               []*cli.Author{{Name: "Tomás Weigenast"}},
		Copyright:             "GPL-3.0 license",
		Name:                  "nexema",
		CustomAppHelpTemplate: helpText,
		EnableBashCompletion:  true,
		Commands:              commands,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func buildCmdExecutor(ctx *cli.Context) error {
	args := ctx.Args()
	inputPath := args.Get(0)
	outputPath := args.Get(1)

	if inputPath == "" {
		return cli.ShowCommandHelp(ctx, "build")
	}

	if outputPath == "" {
		return cli.ShowCommandHelp(ctx, "generbuildate")
	}

	builder := NewBuilder()
	err := builder.Build(inputPath)
	if err != nil {
		return cli.Exit(err.Error(), -1)
	}

	err = builder.Snapshot()
	if err != nil {
		return cli.Exit(err.Error(), -1)
	}

	return nil
}

func generateCmdExecutor(ctx *cli.Context) error {
	args := ctx.Args()
	inputPath := args.Get(0)
	outputPath := args.Get(1)

	if inputPath == "" {
		return cli.ShowCommandHelp(ctx, "generate")
	}

	if outputPath == "" {
		return cli.ShowCommandHelp(ctx, "generate")
	}

	builder := NewBuilder()
	err := builder.Build(inputPath)
	if err != nil {
		return cli.Exit(err.Error(), -1)
	}

	return nil
}

func formatCmdExecutor(ctx *cli.Context) error {
	return nil
}
