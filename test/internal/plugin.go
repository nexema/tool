package internal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

// Plugin specifies the details of a schema_interpreter plugin
// used to generate code.
type Plugin struct {
	Executable string   // The path to the plugin executable
	Args       []string // The arguments to pass to the command
}

// NewPlugin creates a new Plugin with an executable path
func NewPlugin(exe string, args ...string) *Plugin {
	return &Plugin{
		Executable: exe,
		Args:       args,
	}
}

func (p *Plugin) Call(outputType string, output string) error {
	p.Args = append(p.Args, fmt.Sprintf("type:%s", outputType))
	subProcess := exec.Command(p.Executable, p.Args...)

	stdin, err := subProcess.StdinPipe()
	if err != nil {
		return fmt.Errorf("could not get stdin pipe: %s", err)
	}

	stdout, err := subProcess.StdoutPipe()
	if err != nil {
		return fmt.Errorf("could not get stdout pipe: %s", err)
	}

	defer func(stdin io.WriteCloser) {
		err := stdin.Close()
		if err != nil {
			panic(fmt.Errorf("panic: %s", err))
		}
	}(stdin)

	if err = subProcess.Start(); err != nil {
		return fmt.Errorf("could not start subprocess: %s", err)
	}

	// run goroutine to wait for output
	c := make(chan string)
	reader := bufio.NewReader(stdout)
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			c <- scanner.Text()
			break
		}
	}(reader)

	_, err = io.WriteString(stdin, output)
	if err != nil {
		return errors.New("failed to write output")
	}

	_, err = io.WriteString(stdin, "\n")
	if err != nil {
		return errors.New("failed to write new line")
	}

	err = subProcess.Wait()
	if err != nil {
		return err
	}

	outputResult := <-c
	if outputResult != "ok" {
		return fmt.Errorf("plugin does not returned 'ok'. given: %s", outputResult)
	}

	return nil
}

// func (p *Plugin) waitForOk(scanner *bufio.Scanner) error {
// 	fmt.Println("gonna scan")
// 	for {
// 		fmt.Println("waiting")
// 		if scanner.Scan() {
// 			buf := scanner.Bytes()
// 			if string(buf) != "ok" {
// 				fmt.Println("error happened")
// 				return fmt.Errorf("plugin does not returned 'ok'. given: %s", string(buf))
// 			}

// 			break
// 		}
// 	}

// 	return nil
// }
