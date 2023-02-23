package nexema

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	jsoniter "github.com/json-iterator/go"
)

const pluginDiscoverUrl = "https://raw.githubusercontent.com/nexema/.wkp/main/wkp.json"

var (
	nexemaFolder      string
	discoveredPlugins map[string]WellKnownPlugin
	httpClient        *http.Client = &http.Client{}
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

func downloadPlugin(name, pluginPath string) error {
	if discoveredPlugins == nil {
		err := DiscoverWellKnownPlugins()
		if err != nil {
			return err
		}
	}

	wkp, ok := discoveredPlugins[name]
	if !ok {
		return fmt.Errorf("well known plugin %q not found", name)
	}

	resp, err := httpClient.Get(wkp.DownloadUrl)
	if err != nil {
		return fmt.Errorf("could not download plugin %q, error: %s", name, err)
	}
	defer resp.Body.Close()

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		panic(err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	err = os.MkdirAll(pluginPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create output directory for plugin %q, error: %s", name, pluginPath)
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("could not extract plugin %q downloaded contents, error: %s", name, err)
		}

		outfile, err := os.Create(header.Name)
		if err != nil {
			return fmt.Errorf("could not save extracted contents of plugin %q, error: %s", name, err)
		}
		defer outfile.Close()

		_, err = io.Copy(outfile, tr)
		if err != nil {
			return fmt.Errorf("could not copy contents of the file %q from tar, error: %s", name, err)
		}
	}

	return nil
}
