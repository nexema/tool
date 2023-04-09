package nexema

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

const pluginDiscoverUrl = "https://raw.githubusercontent.com/nexema/.wkp/main/wkp.json"

var (
	nexemaFolder      string
	discoveredPlugins map[string]WellKnownPlugin
)

// Run initializes the Nexema tool in the machine, creating the nexema folder in application documents folder if not found
func Run() {
	dir, err := os.UserConfigDir()
	if err != nil {
		panic(fmt.Errorf("could not determine user configuration directory, cannot create .nexema folder. Error: %s", err))
	}

	nexemaFolder = path.Join(dir, "nexema")
	err = os.MkdirAll(nexemaFolder, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("could not create .nexema folder at user configuration directory. Error: %s", err))
	}
}

// DiscoverWellKnownPlugins looks up for Nexema's well known plugin directory and extracts its information
// for later download
func DiscoverWellKnownPlugins() error {
	resp, err := http.Get(pluginDiscoverUrl)
	if err != nil {
		return fmt.Errorf("could not get information of discover url. Error: %s", err)
	}
	defer resp.Body.Close()

	discoveredPlugins = make(map[string]WellKnownPlugin)
	if err := jsoniter.NewDecoder(resp.Body).Decode(&discoveredPlugins); err != nil {
		return fmt.Errorf("could not decode discovered information. Error: %s", err)
	}

	return nil
}

func GetWellKnownPlugins() []PluginInfo {
	out := make([]PluginInfo, len(discoveredPlugins))
	i := 0
	for name, details := range discoveredPlugins {
		out[i] = PluginInfo{Name: name, Version: details.Version}
		i++
	}

	return out
}

// GetWellKnownPlugin returns the path to the wkp name. If cannot find it, it will try to download it using DiscoverWellKnownPlugins
func GetWellKnownPlugin(name string) (pluginPath string, err error) {
	pluginPath = path.Join(nexemaFolder, "plugins", name)
	_, err = os.Stat(pluginPath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// try to download
		err = downloadPlugin(name, pluginPath)
		if err != nil {
			return "", err
		}
	}

	pluginPath = path.Join(pluginPath, name)
	return
}

func GetInstalledPlugins() []PluginInfo {
	installed := make([]PluginInfo, 0)

	buffer, err := os.ReadFile(path.Join(nexemaFolder, "plugins", ".plugins"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return installed
		}

		panic(err)
	}

	lines := strings.Split(string(buffer), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			panic("invalid .plugins file")
		}

		installed = append(installed, PluginInfo{parts[0], parts[1]})
	}

	return installed
}

func downloadPlugin(name, pluginPath string) error {
	if discoveredPlugins == nil {
		err := DiscoverWellKnownPlugins()
		if err != nil {
			return err
		}
	}

	// get the plugin info
	wkp, ok := discoveredPlugins[name]
	if !ok {
		return fmt.Errorf("well known plugin %q not found", name)
	}

	// create an output folder for it
	err := os.Mkdir(path.Join(nexemaFolder, "plugins", name), os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create output folder for plugin %q", name)
	}

	for _, step := range wkp.InstallSteps {
		step = strings.ReplaceAll(step, "%packageName%", wkp.PackageName)
		step = strings.ReplaceAll(step, "%version%", wkp.Version)

		toks := strings.Split(step, " ")
		commandName := toks[0]
		args := toks[0:]

		cmd := exec.Command(commandName, args...)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("could not execute command %q, error: %s", step, err)
		}

		fmt.Printf("[nexema] ran %q output -> %s\n", step, output)
	}

	// once installed, write to .plugins
	f, err := os.OpenFile(path.Join(nexemaFolder, "plugins", ".plugins"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("fatal error -> could not open .plugins file, error: %s", err)
	}

	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("%s:%s", name, wkp.Version)); err != nil {
		return fmt.Errorf("fatal error -> could not write to .plugins file, error: %s", err)
	}

	return nil
}
