package cmd

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"
	nexema "tomasweigenast.com/nexema/tool/internal/nexema"
)

func pluginList(ctx *cli.Context) error {
	plugins := nexema.GetInstalledPlugins()

	if len(plugins) == 0 {
		fmt.Println(`There are not any installed Nexema's plugins. List installable plugins with "nexema plugin discover".`)
		return nil
	}

	fmt.Printf("=========================\nInstalled plugins\n\n")
	for _, info := range plugins {
		fmt.Printf("    %s: v%s\n", info.Name, info.Version)
	}
	fmt.Println("=========================")
	return nil
}

func pluginDiscover(ctx *cli.Context) error {
	err := nexema.DiscoverWellKnownPlugins()
	if err != nil {
		return err
	}

	plugins := nexema.GetWellKnownPlugins()
	fmt.Printf("=========================\nNexema available plugins\n\n")
	for name, info := range *plugins {
		fmt.Printf("    %s: v%s\n", name, info.Version)
	}
	fmt.Println("=========================")

	return nil
}

func pluginInstall(ctx *cli.Context) error {
	pluginName := ctx.Args().Get(0)
	if len(pluginName) == 0 {
		return fmt.Errorf("plugin name is required")
	}

	err := nexema.InstallPlugin(pluginName)
	if err != nil {
		if errors.Is(err, nexema.ErrPluginUpgrade) {
			return fmt.Errorf("plugin needs an upgrade, run nexema plugin upgrade %s", pluginName)
		}
	}

	return err
}

func pluginUpgrade(ctx *cli.Context) error {
	pluginName := ctx.Args().Get(0)
	if len(pluginName) == 0 {
		return fmt.Errorf("plugin name is required")
	}

	return nexema.UpgradePlugin(pluginName)
}
