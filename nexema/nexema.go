package nexema

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
)

const (
	pluginDiscoverUrl = "https://raw.githubusercontent.com/nexema/.wkp/main/wkp.json"
	version           = "1.0.0"
	configFileName    = ".nexema"
)

var singleton *Nexema = newNexema()

// Nexema handles everything about a Nexema execution instance
type Nexema struct {
	nexemaFolder  string
	pluginsFolder string

	discoveredPlugins *map[string]WellKnownPlugin
	config            *NexemaConfig
	configFile        *os.File
	logFile           *os.File
}

// NexemaConfig contains information about the installed Nexema binary
type NexemaConfig struct {
	Version          string                `json:"version"`   // The installed Nexema version
	InstalledPlugins map[string]PluginInfo `json:"installed"` // The list of installed plugins
}

func newNexema() *Nexema {
	dir, err := os.UserConfigDir()
	if err != nil {
		panic(fmt.Errorf("could not determine user configuration directory, cannot create nexema folder. Error: %s", err))
	}

	nexemaFolder := path.Join(dir, "nexema")

	return &Nexema{
		nexemaFolder:  nexemaFolder,
		pluginsFolder: path.Join(nexemaFolder, "plugins"),
	}
}

// Run initializes the Nexema tool in the machine, creating the nexema folder in application documents folder if not found
func Run() error {
	singleton.initLogger()

	log.Debugf("Ensuring %s folder exists", singleton.nexemaFolder)

	err := os.MkdirAll(singleton.pluginsFolder, os.ModePerm) // this will create nexema folder too

	if err != nil {
		return fmt.Errorf("could not create plugins folder at user configuration directory. Error: %s", err)
	}

	// read config if exists
	singleton.configFile, err = os.OpenFile(path.Join(singleton.nexemaFolder, configFileName), os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Could not open nexema config file: %s", err)
	}

	return singleton.readConfig()
}

// Closes any opened file
func Exit() {
	if singleton.configFile != nil {
		singleton.configFile.Close()
		singleton.logFile.Close()
	}
}

// DiscoverWellKnownPlugins looks up for Nexema's well known plugin directory and extracts its information
// for later download
func DiscoverWellKnownPlugins() error {
	singleton.discoveredPlugins = new(map[string]WellKnownPlugin)

	resp, err := http.Get(pluginDiscoverUrl)
	if err != nil {
		return fmt.Errorf("could not get information of discover url. Error: %s", err)
	}
	defer resp.Body.Close()

	if err := jsoniter.NewDecoder(resp.Body).Decode(singleton.discoveredPlugins); err != nil {
		return fmt.Errorf("could not decode discovered information. Error: %s", err)
	}
	// jsoniter.UnmarshalFromString(`{"js":{"version":"1.0.7","packageName":"nexema-generator","steps":["npm pack %packageName%@%version%","npm install -g %packageName%-%version%.tgz --prefix=.", "rm %packageName%-%version%.tgz"],"binary":"bin/nexemajsgen"}}`, singleton.discoveredPlugins)

	return nil
}

// GetWellKnownPlugins returns the list of well known Nexema plugins
func GetWellKnownPlugins() map[string]WellKnownPlugin {
	if singleton.discoveredPlugins == nil {
		err := DiscoverWellKnownPlugins()
		if err != nil {
			panic(err)
		}
	}

	return *singleton.discoveredPlugins
}

// GetWellKnownPlugin returns the path to  a well known Nexema plugin. If its not installed, it will install it first.
func GetWellKnownPlugin(name string) (pluginPath string, err error) {
	if plugin, ok := singleton.config.InstalledPlugins[name]; ok {
		pluginPath = plugin.Path
	} else {
		err = InstallPlugin(name)
		if err != nil {
			return "", err
		}

		pluginPath = singleton.config.InstalledPlugins[name].Path
	}

	return
}

// GetPluginInfo returns information about an installed Nexema plugin
func GetPluginInfo(name string) *PluginInfo {
	info, ok := singleton.config.InstalledPlugins[name]
	if ok {
		return &info
	}

	return nil
}

// InstallPlugin installs a Nexema plugin
func InstallPlugin(name string) error {
	pluginInfo := GetPluginInfo(name)
	if pluginInfo != nil {
		return fmt.Errorf("Plugin %s already installed", name)
	}

	err := downloadPlugin(name)
	if err != nil {
		return err
	}

	return nil
}

// GetInstalledPlugins returns the list of installed Nexema, well known plugins.
func GetInstalledPlugins() map[string]PluginInfo {
	return singleton.config.InstalledPlugins
}

func downloadPlugin(name string) error {
	log.Debug("Downloading plugin ", name)

	if singleton.discoveredPlugins == nil {
		err := DiscoverWellKnownPlugins()
		if err != nil {
			return err
		}
	}

	// get the well known plugin info
	wkp, ok := (*singleton.discoveredPlugins)[name]
	if !ok {
		return fmt.Errorf("well known plugin %q not found", name)
	}

	// create an output folder for it
	pluginFolder := path.Join(singleton.nexemaFolder, "plugins", name)
	err := os.MkdirAll(pluginFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create output folder for plugin %q, error: %s", name, err)
	}

	log.Debug("Running installation steps...")
	for _, step := range wkp.InstallSteps {
		step = strings.ReplaceAll(step, "%packageName%", wkp.PackageName)
		step = strings.ReplaceAll(step, "%version%", wkp.Version)
		step = strings.ReplaceAll(step, "%pluginFolder%", pluginFolder)

		log.Debug("Going to run: ", step)

		toks := strings.Split(step, " ")

		cmd := exec.Command("sudo", toks...)
		cmd.Dir = pluginFolder
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			os.RemoveAll(pluginFolder)
			return fmt.Errorf("could not execute command %q, error: %s", step, err)
		}

		log.Debugf("[nexema] ran %q output -> %s\n", step, output)
	}

	log.Debugln("Plugin installed. Updating config")

	singleton.config.InstalledPlugins[name] = PluginInfo{
		Name:    name,
		Version: wkp.Version,
		Path:    path.Join(pluginFolder, wkp.BinaryName),
	}
	singleton.writeConfig()

	log.Debugln("Plugin installed succesfully")

	return nil
}

func (n *Nexema) initLogger() {
	logFile := path.Join(n.nexemaFolder, "log.txt")
	var err error
	n.logFile, err = os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("failed to create logfile (%s): %s", logFile, err))
	}
	log.SetOutput(n.logFile)
	log.SetOutput(os.Stdout)
}

func (n *Nexema) readConfig() error {
	scanner := bufio.NewScanner(n.configFile)
	scanned := scanner.Scan()
	buffer := scanner.Bytes()
	if !scanned {
		// write default config
		n.config = &NexemaConfig{
			Version:          version,
			InstalledPlugins: make(map[string]PluginInfo),
		}
		return n.writeConfig()
	} else {
		n.config = &NexemaConfig{}
		err := jsoniter.Unmarshal(buffer, n.config)
		if err != nil {
			return fmt.Errorf("could not read nexema config file, err: %s", err)
		}

		return n.verifyVersion()
	}
}

func (n *Nexema) writeConfig() error {
	buffer, err := jsoniter.Marshal(n.config)
	if err != nil {
		return fmt.Errorf("could not serialize nexema config, error: %s", err)
	}

	n.configFile.Seek(0, 0)
	_, err = n.configFile.Write(buffer)
	if err != nil {
		return fmt.Errorf("could not write config to file, error: %s", err)
	}

	return nil
}

func (n *Nexema) verifyVersion() error {
	if n.config.Version != version {
		n.config.Version = version
		return n.writeConfig()
	}
	return nil
}
