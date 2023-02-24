package plugin

import (
	"bytes"
	"os/exec"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// Plugin contains information about a Nexema generator plugin
type Plugin struct {
	Name    string // the name of the plugin
	BinPath string // the path to the binary executable
}

func NewPlugin(name, binPath string) *Plugin {
	return &Plugin{name, binPath}
}

// Run sends blob to the plugin stdin and returns the output of the plugin
func (p *Plugin) Run(blob []byte, arguments []string) (*PluginResult, error) {
	blob = append(blob, '\n') // always append a new line
	buffer := bytes.NewBuffer(blob)
	respBuffer := new(bytes.Buffer)

	cmd := exec.Command(p.BinPath, arguments...)
	cmd.Stdin = buffer
	cmd.Stdout = respBuffer

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	resp := strings.TrimSpace(respBuffer.String())
	result := &PluginResult{}
	err = jsoniter.UnmarshalFromString(resp, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
