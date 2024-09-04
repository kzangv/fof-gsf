//go:build !windows

package shell

import (
	"errors"
	"os/exec"
	"syscall"
)

type UnixShell struct {
	*exec.Cmd
}

func (c *UnixShell) New(command string) {
	c.Cmd = exec.Command("/bin/bash", "-c", command)
	c.Cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func (c UnixShell) Exec() (string, error) {
	output, err := c.Cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (c UnixShell) Destroy() error {
	if c.Cmd.Process.Pid > 0 {
		return syscall.Kill(-c.Cmd.Process.Pid, syscall.SIGKILL)
	}
	return errors.New("timeout killed")
}

func newCmd() Cmd {
	return &UnixShell{}
}
