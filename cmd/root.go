package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	nexema "tomasweigenast.com/nexema/tool/internal/nexema"
)

const helpText = `Nexema - binary interchange made simple

Usage:
   nexema [command] (arguments...)
{{if .Commands}}
Available commands:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}{{end}}
Available flags:
{{range .VisibleFlags}}   {{.}}
{{end}}
About:
	Made by Tomás Weigenast <tomaswegenast@gmail.com>
	v1.0.0
	Licensed under GPL-3.0
`

var app *cli.Command

func init() {

	app = &cli.Command{
		Name:                          "nexema",
		CustomRootCommandHelpTemplate: helpText,
		CommandNotFound: func(ctx *cli.Context, s string) {
			cli.ShowAppHelp(ctx)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:       "verbose",
				Required:   false,
				Hidden:     false,
				Persistent: true,
				Action: func(ctx *cli.Context, b bool) error {
					fmt.Println("verbose logging enabled")
					logrus.SetLevel(logrus.TraceLevel)
					return nil
				},
				Usage: "Print all logs to the console",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "mod",
				Usage: "Manages Nexema projects",
				Commands: []*cli.Command{
					{
						Name:      "init",
						Usage:     "Initializes a new project",
						ArgsUsage: "[the path where to initialize the project]",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:     "overwrite",
								Usage:    "Overwrites any previous existing nexema.yaml at specified at",
								Required: false,
							},
						},
						Action: modInit,
					},
					{
						Name:  "generator",
						Usage: "Configures the generators of the plugin",
						Commands: []*cli.Command{
							{
								Name:      "add",
								Usage:     "Adds a new generator",
								ArgsUsage: "[plugin name]",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:     "bin-path",
										Usage:    "The path to the plugin executable, if it's a non well known plugin",
										Required: false,
										Hidden:   false,
									},
								},
								Action: modAddGenerator,
							},
							{
								Name:      "remove",
								Usage:     "Removes a generator",
								ArgsUsage: "[plugin name]",
								Action:    modRemoveGenerator,
							},
						},
					},
				},
			},
			{
				Name:  "build",
				Usage: "Builds a project and optionally outputs a snapshot file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "out",
						Usage: "The path to the output folder where to write the snapshot file",
					},
				},
				Action: build,
			},
			{
				Name:  "generate",
				Usage: "Builds a project and generates source code",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "snapshot-file",
						Usage: "generate from a snapshot file",
					},
					&cli.StringSliceFlag{
						Required: true,
						Name:     "for",
						Usage:    "the generators to use and their output path",
					},
				},
				Action: generate,
			},
			{
				Name:  "format",
				Usage: "Format all .nex files in the specified project",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "path",
						Usage: "Path to the project directory",
					},
				},
				Action: format,
			},
			{
				Name:  "plugin",
				Usage: "Manage Nexema plugins",
				Action: func(ctx *cli.Context) error {
					return cli.ShowSubcommandHelp(ctx)
				},
				Commands: []*cli.Command{
					{
						Name:   "list",
						Usage:  "List installed Nexema plugins",
						Action: pluginList,
					},
					{
						Name:   "discover",
						Usage:  "List all well-known Nexema plugins",
						Action: pluginDiscover,
					},
					{
						Name:      "install",
						Usage:     "Installs a Nexema well-known plugin",
						ArgsUsage: "[plugin-name]",
						Action:    pluginInstall,
					},
					{
						Name:      "upgrade",
						Usage:     "Upgrades a installed Nexema well-known plugin",
						ArgsUsage: "[plugin-name]",
						Action:    pluginUpgrade,
					},
				},
			},
			{
				Name:  "config",
				Usage: "Shows the path to the Nexema configuration",
				Action: func(ctx *cli.Context) error {
					path := nexema.GetNexemaFolder()
					fmt.Printf("Nexema config is located at: %s\n", path)
					return nil
				},
			},
		},
	}
}

func Execute() {
	nexema.Init()

	logrus.SetLevel(logrus.TraceLevel)

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
	}

	nexema.Exit()
}
