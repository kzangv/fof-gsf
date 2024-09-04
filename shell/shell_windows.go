//go:build windows

package shell

import (
	"errors"
	"os/exec"
	"syscall"
)

type WindowShell struct {
	cmd *exec.Cmd
}

func (c *WindowShell) New(command string) {
	c.Cmd = exec.Command("cmd", "/C", command)
	// 隐藏cmd窗口
	c.Cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}

func (c WindowShell) Exec() (string, error) {
	var (
		output []byte
		err    error
	)
	if output, err = c.Cmd.CombinedOutput(); err != nil {
		return "", err
	}
	reader := transform.NewReader(bytes.NewReader(output), simplifiedchinese.GBK.NewDecoder())
	if output, err = ioutil.ReadAll(reader); err != nil {
		return "", err
	}
	return string(output), nil
}

func (c WindowShell) Destroy() error {
	if c.Cmd.Process.Pid > 0 {
		exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
		return cmd.Process.Kill()
	}
	return errors.New("timeout killed")
}

func newCmd() Cmd {
	return &WindowShell{}
}
