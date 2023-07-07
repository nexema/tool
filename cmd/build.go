package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func build(ctx *cli.Context) error {
	path := ctx.Args().First()
	if len(path) == 0 {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	// outputPath := ctx.String("out")
	// return buildCmd(path, outputPath)
	return nil
}

func generate(ctx *cli.Context) error {
	path := ctx.Args().First()
	snapshotPath := ctx.String("snapshot-file")
	if len(path) == 0 {
		return cli.Exit("path is required", 1)
	}

	generateFor := ctx.StringSlice("for")

	_ = snapshotPath
	_ = generateFor

	// return generateCmd(path, snapshotPath, generateFor)
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
