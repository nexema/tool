package plugin

import (
	"bytes"
	"os/exec"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

// Plugin contains information about a Nexema generator plugin
type Plugin struct {
	Name    string // the name of the plugin
	BinPath string // the path to the binary executable
}

func NewPlugin(name, binPath string) *Plugin {
	return &Plugin{name, binPath}
}

// Run sends blob to the plugin stdin and returns the output of the plugin.
//
// Plugin's options are passed as initial arguments
func (p *Plugin) Run(blob []byte, arguments []string, env []string) (*PluginResult, error) {
	blob = append(blob, '\n') // always append a new line
	buffer := bytes.NewBuffer(blob)
	respBuffer := new(bytes.Buffer)

	cmd := exec.Command(p.BinPath, arguments...)
	cmd.Stdin = buffer
	cmd.Stdout = respBuffer
	cmd.Stderr = respBuffer
	cmd.Env = env
	err := cmd.Run()

	logrus.Tracef("Command %s (args: %s) ran. Error: %s\n", p.BinPath, arguments, err)
	if respBuffer.Len() != 0 {
		logrus.Tracef("Response buffer: %s\n", respBuffer.String())
	}

	if err != nil {
		return nil, err
	}

	result := &PluginResult{}
	err = jsoniter.Unmarshal(respBuffer.Bytes(), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
