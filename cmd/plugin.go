package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Plugin represents the lang-side of the generation.
type Plugin struct {
	Name     string
	ExecPath string // The path to the executable or its main entry point if added to PATH
}

// NewPlugin creates a new Plugin instance.
func NewPlugin(name, execPath string) *Plugin {
	return &Plugin{Name: name, ExecPath: execPath}
}

// Run sends blob to the plugin stdin and returns nil if it was successful; the error otherwise.
func (p *Plugin) Run(blob []byte) error {
	blob = append(blob, '\n') // always append a new line
	buffer := bytes.NewBuffer(blob)
	respBuffer := new(bytes.Buffer)

	cmd := exec.Command(p.ExecPath)
	cmd.Stdin = buffer
	cmd.Stdout = respBuffer

	err := cmd.Run()
	if err != nil {
		return err
	}

	resp := strings.TrimSpace(respBuffer.String())
	if resp == "ok" {
		return nil
	}

	return fmt.Errorf("an error occurred trying to generate source code for %s. %s", p.Name, resp)
}
