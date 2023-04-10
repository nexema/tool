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
	log "github.com/sirupsen/logrus"
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
	log.Debugf("Ensuring %s folder exists", nexemaFolder)
	err = os.MkdirAll(nexemaFolder, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("could not create nexema folder at user configuration directory. Error: %s", err))
	}

	pluginsFolder := path.Join(nexemaFolder, "plugins")
	err = os.MkdirAll(pluginsFolder, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("could not create plugins folder at user configuration directory. Error: %s", err))
	}

	initLogger()
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

// GetPluginInfo returns information about an installed Nexema plugin
func GetPluginInfo(name string) *PluginInfo {
	pluginPath := path.Join(nexemaFolder, "plugins", name)
	_, err := os.Stat(pluginPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		panic(err)
	}

	return &PluginInfo{
		Path:    pluginPath,
		Name:    name,
		Version: "1",
	}
}

// InstallPlugin installs a Nexema plugin
func InstallPlugin(name string) error {
	pluginInfo := GetPluginInfo(name)
	if pluginInfo != nil {
		return fmt.Errorf("Plugin %s already installed", name)
	}

	pluginPath := path.Join(nexemaFolder, "plugins", name)
	err := downloadPlugin(name, pluginPath)
	if err != nil {
		return err
	}

	return nil
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

		installed = append(installed, PluginInfo{parts[0], parts[1], ""})
	}

	return installed
}

func downloadPlugin(name, pluginPath string) error {
	log.Debug("Downloading plugin ", name)

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
	pluginFolder := path.Join(nexemaFolder, "plugins", name)
	err := os.Mkdir(pluginFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create output folder for plugin %q, error: %s", name, err)
	}

	log.Debug("Running installation steps...")
	for _, step := range wkp.InstallSteps {
		step = strings.ReplaceAll(step, "%packageName%", wkp.PackageName)
		step = strings.ReplaceAll(step, "%version%", wkp.Version)
		step = strings.ReplaceAll(step, "%outputFolder%", pluginFolder)

		log.Debug("Going to run: ", step)

		toks := strings.Split(step, " ")
		// commandName := toks[0]
		// args := toks[0:]

		cmd := exec.Command("sudo", toks...)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			os.RemoveAll(pluginFolder)
			return fmt.Errorf("could not execute command %q, error: %s", step, err)
		}

		log.Debugf("[nexema] ran %q output -> %s\n", step, output)
	}

	log.Debugln("Plugin installed. Updating .plugins file")

	// once installed, write to .plugins
	f, err := os.OpenFile(path.Join(nexemaFolder, "plugins", ".plugins"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("fatal error -> could not open .plugins file, error: %s", err)
	}

	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("%s:%s", name, wkp.Version)); err != nil {
		return fmt.Errorf("fatal error -> could not write to .plugins file, error: %s", err)
	}

	log.Debugln("Plugin installed succesfully")

	return nil
}

func initLogger() {
	// logFile := path.Join(nexemaFolder, fmt.Sprintf("log_%d.txt", time.Now().Unix()))
	// f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	// if err != nil {
	// 	panic(fmt.Errorf("failed to create logfile (%s): %s", logFile, err))
	// }

	// defer f.Close()

	// log.SetOutput(f)
	log.SetOutput(os.Stdout)
}
