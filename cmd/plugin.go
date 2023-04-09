package cmd

import (
	"fmt"

	"github.com/urfave/cli"
	"tomasweigenast.com/nexema/tool/nexema"
)

func pluginInstall(c *cli.Context) error {
	pluginName := c.Args().Get(0)
	if len(pluginName) == 0 {
		return fmt.Errorf("plugin name is required")
	}

	nexema.GetWellKnownPlugin(pluginName)
	return nil
}

func pluginList(c *cli.Context) error {
	plugins := nexema.GetInstalledPlugins()
	fmt.Printf("=========================\nInstalled plugins\n\n")
	for _, info := range plugins {
		fmt.Printf("    %s: v%s\n", info.Name, info.Version)
	}
	fmt.Println("=========================")
	return nil
}

func pluginDiscover(c *cli.Context) error {
	err := nexema.DiscoverWellKnownPlugins()
	if err != nil {
		return err
	}

	plugins := nexema.GetWellKnownPlugins()
	fmt.Printf("=========================\nNexema available plugins\n\n")
	for _, info := range plugins {
		fmt.Printf("    %s: v%s\n", info.Name, info.Version)
	}
	fmt.Println("=========================")

	return nil
}
