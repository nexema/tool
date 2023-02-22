package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

const helpText = `Nexema - binary interchange made simple

Usage:
   nexema command (arguments...)
{{if .Commands}}
Available commands:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
{{end}}
About:
	Made by Tomás Weigenast <tomaswegenast@gmail.com>
	v1.0.0
	Licensed under GPL-3.0
`

var app *cli.App

func init() {
	app = &cli.App{
		CustomAppHelpTemplate: helpText,
		CommandNotFound:       cli.ShowCommandCompletions,
	}

	app.Commands = []cli.Command{
		{
			Name:  "mod",
			Usage: "Manages Nexema projects",
			Subcommands: []cli.Command{
				{
					Name:      "init",
					Usage:     "Initializes a new project",
					ArgsUsage: "[the path where to initialize the project]",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:     "overwrite",
							Usage:    "Overwrites any previous existing nexema.yaml at specified at",
							Required: false,
						},
					},
					Action: func(c *cli.Context) error {
						path := c.Args().First()
						if len(path) == 0 {
							return cli.NewExitError("path is required", 1)
						}

						return modInit(path, c.Bool("overwrite"))
					},
				},
			},
		},
		{
			Name:  "build",
			Usage: "Builds a project",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "path",
					Usage: "Path to the project directory",
				},
			},
			Action: func(c *cli.Context) error {
				path := c.String("path")
				if path == "" {
					return cli.NewExitError("path is required", 1)
				}
				fmt.Printf("Building the project at %s...\n", path)
				return nil
			},
		},
		{
			Name:  "generate",
			Usage: "Builds a project and generates source code",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "path",
					Usage: "Path to the project directory",
				},
			},
			Action: func(c *cli.Context) error {
				path := c.String("path")
				if path == "" {
					return cli.NewExitError("path is required", 1)
				}
				fmt.Printf("Generating code for the project at %s...\n", path)
				return nil
			},
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
			Action: func(c *cli.Context) error {
				path := c.String("path")
				if path == "" {
					return cli.NewExitError("path is required", 1)
				}
				fmt.Printf("Formatting code for the project at %s...\n", path)
				return nil
			},
		},
	}
}

func Execute() {
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
