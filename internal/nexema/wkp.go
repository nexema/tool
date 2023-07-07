package nexema

import "github.com/Masterminds/semver/v3"

type WellKnownPlugin struct {
	Version      semver.Version `json:"version"`
	PackageName  string         `json:"packageName"`
	InstallSteps []string       `json:"steps"`
	BinaryName   string         `json:"binary"`
}

type PluginInfo struct {
	Name    string         `json:"name"`    // the name of the plugin
	Version semver.Version `json:"version"` // the installed version
	Path    string         `json:"path"`    // the path to the binary
}
