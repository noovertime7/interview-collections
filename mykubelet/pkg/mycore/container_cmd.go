package mycore

import "os/exec"

type ContainerCmd struct {
	Cmd           *exec.Cmd `json:"cmd"`
	ContainerName string    `json:"container_name"`
	ExitCode      int       `json:"exit_code"`
	StdError      error     `json:"std_error"`
	StdOutput     string    `json:"std_output"`
}

func (c *ContainerCmd) Run() {
	output, err := c.Cmd.Output()
	c.StdOutput = string(output)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			c.ExitCode = ee.ExitCode()
			c.StdError = ee
		} else {
			c.ExitCode = -9999
			c.StdError = err
		}
	}
}
