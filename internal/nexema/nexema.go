package nexema

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const (
	pluginDiscoverUrl = "https://raw.githubusercontent.com/nexema/.wkp/main/wkp.json"
	version           = "1.0.0"
	configFileName    = ".nexema"
)

type (
	// Nexema handles everything about a Nexema execution instance
	Nexema struct {
		nexemaFolder  string
		pluginsFolder string

		discoveredPlugins *map[string]WellKnownPlugin
		config            *NexemaConfig

		configFile *os.File
		logFile    *os.File
	}

	// NexemaConfig contains information about the installed Nexema binary
	NexemaConfig struct {
		Version          string                `json:"version"`   // The installed Nexema version
		InstalledPlugins map[string]PluginInfo `json:"installed"` // The list of installed plugins
	}
)

var (
	singleton *Nexema

	ErrPluginUpgrade error = errors.New("plugin upgrade")
)

func newNexema() (*Nexema, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine user configuration's directory to create the nexema folder, error: %v", err)
	}

	nexemaFolder := path.Join(dir, "nexema")
	nexema := &Nexema{
		nexemaFolder:      nexemaFolder,
		pluginsFolder:     path.Join(nexemaFolder, "plugins"),
		discoveredPlugins: new(map[string]WellKnownPlugin),
		config:            &NexemaConfig{},
	}

	err = nexema.initLogger()
	if err != nil {
		return nil, err
	}

	err = nexema.initConfig()
	if err != nil {
		return nil, err
	}

	return nexema, nil
}

func (self *Nexema) initLogger() error {
	logFile := path.Join(self.nexemaFolder, "log.txt")
	var err error

	self.logFile, err = os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create log file, error: %v", err)
	}

	log.SetOutput(self.logFile)
	log.SetOutput(os.Stdout)
	return nil
}

func (self *Nexema) initConfig() error {
	log.Debugf("Ensuring %s folder exists", self.nexemaFolder)

	err := os.MkdirAll(self.pluginsFolder, os.ModePerm) // this will create nexema folder too
	if err != nil {
		return fmt.Errorf("could not create plugins folder, error: %s", err)
	}

	// read config if exists
	self.configFile, err = os.OpenFile(path.Join(self.nexemaFolder, configFileName), os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Could not open config file, error: %s", err)
	}

	scanner := bufio.NewScanner(self.configFile)
	scanned := scanner.Scan()
	buffer := scanner.Bytes()
	if scanned {
		self.config = &NexemaConfig{}
		err := jsoniter.Unmarshal(buffer, self.config)
		if err != nil {
			panic(fmt.Errorf("could not read nexema config from file, error: %s", err))
		}

		self.verifyVersion()
	} else {
		// write default config
		self.config = &NexemaConfig{
			Version:          version,
			InstalledPlugins: make(map[string]PluginInfo),
		}

		self.writeConfig()
	}

	return nil
}

func (self *Nexema) writeConfig() {
	buffer, err := jsoniter.Marshal(self.config)
	if err != nil {
		panic(fmt.Errorf("could not serialize nexema config, error: %s", err))
	}

	self.configFile.Seek(0, 0)
	_, err = self.configFile.Write(buffer)
	if err != nil {
		panic(fmt.Errorf("could not write config to file, error: %s", err))
	}
}

func (self *Nexema) verifyVersion() {
	if self.config.Version != version {
		self.config.Version = version
		self.writeConfig()
	}
}

// Init initializes the Nexema singleton
func Init() error {
	var err error
	singleton, err = newNexema()
	return err
}

// Exit closes any opened file
func Exit() {
	if singleton.configFile != nil {
		singleton.configFile.Close()
	}

	if singleton.logFile != nil {
		singleton.logFile.Close()
	}
}

// DiscoverWellKnownPlugins looks up for Nexema's well known repository and extracts its information
// for later usage.
func DiscoverWellKnownPlugins() error {
	if singleton.discoveredPlugins != nil && len(*singleton.discoveredPlugins) > 0 {
		return nil
	}

	resp, err := http.Get(pluginDiscoverUrl)
	if err != nil {
		return fmt.Errorf("could not get information of discover url, error: %s", err)
	}

	defer resp.Body.Close()

	if err := jsoniter.NewDecoder(resp.Body).Decode(singleton.discoveredPlugins); err != nil {
		return fmt.Errorf("could not decode discovered information. Error: %s", err)
	}

	return nil
}

// GetWellKnownPlugins returns the list of well known Nexema plugins
func GetWellKnownPlugins() *map[string]WellKnownPlugin {
	if singleton.discoveredPlugins == nil {
		err := DiscoverWellKnownPlugins()
		if err != nil {
			panic(err)
		}
	}

	return singleton.discoveredPlugins
}

// GetWellKnownPlugin returns the path to a well known Nexema plugin. If its not installed, it will install it first.
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
		err := DiscoverWellKnownPlugins()
		if err != nil {
			return err
		}

		wkpVersion := (*singleton.discoveredPlugins)[name].Version
		logrus.Debugf("Current plugin version [%s] - New version [%s]\n", pluginInfo.Version, wkpVersion)
		if pluginInfo.Version.LessThan(&wkpVersion) {
			return ErrPluginUpgrade
		}

		return fmt.Errorf("Plugin %q already installed", name)
	}

	err := downloadPlugin(name, true)
	if err != nil {
		return err
	}

	return nil
}

// UpgradePlugin upgrades an already installed Nexema plugin
func UpgradePlugin(name string) error {
	pluginInfo := GetPluginInfo(name)
	if pluginInfo == nil {
		return fmt.Errorf("plugin %q is not installed", name)
	}

	err := DiscoverWellKnownPlugins()
	if err != nil {
		return err
	}

	wkpVersion := (*singleton.discoveredPlugins)[name].Version
	if pluginInfo.Version.LessThan(&wkpVersion) {
		return downloadPlugin(name, true)
	}

	return fmt.Errorf("plugin %q has already the latest version", name)
}

// GetInstalledPlugins returns the list of installed Nexema, well known plugins.
func GetInstalledPlugins() map[string]PluginInfo {
	return singleton.config.InstalledPlugins
}

// GetNexemaFolder returns the path to the Nexema folder
func GetNexemaFolder() string {
	return singleton.nexemaFolder
}

func downloadPlugin(name string, isUpgrade bool) error {
	log.Debug("Downloading plugin ", name)

	err := DiscoverWellKnownPlugins()
	if err != nil {
		return err
	}

	// get the well known plugin info
	wkp, ok := (*singleton.discoveredPlugins)[name]
	if !ok {
		return fmt.Errorf("well known plugin %q not found", name)
	}

	// create an output folder for it
	pluginFolder := path.Join(singleton.pluginsFolder, name)
	if isUpgrade {
		err = os.RemoveAll(pluginFolder)
		if err != nil {
			return fmt.Errorf("could not remove plugin %q folder before upgrading, error: %s", name, err)
		}
	}

	err = os.MkdirAll(pluginFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create output folder for plugin %q, error: %s", name, err)
	}

	log.Debug("Running installation steps...")
	for _, step := range wkp.InstallSteps {
		step = strings.ReplaceAll(step, "%packageName%", wkp.PackageName)
		step = strings.ReplaceAll(step, "%version%", wkp.Version.String())
		step = strings.ReplaceAll(step, "%pluginFolder%", pluginFolder)

		log.Debug("Going to run: ", step)

		toks := strings.Split(step, " ")

		var cmd *exec.Cmd

		// if runtime.GOOS == "windows" {
		// 	cmd = exec.Command(toks[0], toks[1:]...)
		// } else {
		// 	cmd = exec.Command("sudo", toks...)
		// }
		cmd = exec.Command(toks[0], toks[1:]...)
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

	log.Infoln("Plugin installed succesfully")

	return nil
}
