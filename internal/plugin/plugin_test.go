package plugin

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPluginRun(t *testing.T) {
	plugin := NewPlugin("test", os.Args[0])
	buffer := make([]byte, 0)
	args := []string{
		"-test.run=TestExecCommandHelper",
	}

	result, err := plugin.Run(buffer, args, getHelperEnv(`{
		"exitCode": 0,
		"error": null,
		"files": [
			{
				"id": "12345",
				"name": "identity/user.nex.ts",
				"contents": ""
			}
		]
	}`, 0))

	require.NoError(t, err)
	require.Equal(t, &PluginResult{
		ExitCode: 0,
		Files: &[]GeneratedFile{
			{
				Id:       "12345",
				Name:     "identity/user.nex.ts",
				Contents: "",
			},
		},
	}, result)

}

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

func getHelperEnv(stdout string, exitCode int) []string {
	return []string{
		"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + stdout,
		"EXIT_STATUS=" + fmt.Sprint(exitCode),
	}
}
