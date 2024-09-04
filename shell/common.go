package shell

import (
	"context"
)

type Result struct {
	output string
	err    error
}

type Cmd interface {
	New(command string)
	Exec() (string, error)
	Destroy() error
}

// ExecShell 执行shell命令，可设置执行超时时间
func ExecShell(ctx context.Context, command string) (string, error) {
	return ExecShellEx(ctx, newCmd(), command)
}

// ExecShellEx 执行shell命令，可设置执行超时时间
func ExecShellEx(ctx context.Context, cmd Cmd, command string) (string, error) {
	cmd.New(command)
	receiver := make(chan Result)
	go func() {
		output, err := cmd.Exec()
		receiver <- Result{output, err}
	}()
	select {
	case <-ctx.Done():
		return "", cmd.Destroy()
	case result := <-receiver:
		return result.output, result.err
	}
}
