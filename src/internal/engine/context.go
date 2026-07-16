package engine

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type RunContext struct {
	Registry     *Registry
	ShellCmd     *exec.Cmd
	Stdin        io.WriteCloser
	Stdout       io.ReadCloser
	StdoutReader *bufio.Reader
	Cwd          string
	EnvFile      string
	waitDone     chan struct{}
}

func NewRunContext(mdPath string) (*RunContext, error) {
	shell, _ := getShell()
	cmd := exec.Command(shell)

	// Create a single pipe for both stdout and stderr
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	cmd.Stdout = pw
	cmd.Stderr = pw

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// We need to wait for the command to finish so we can close the write end of the pipe
	waitDone := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		pw.Close()
		close(waitDone)
	}()

	var cwd string
	if filepath.IsAbs(mdPath) {
		cwd = filepath.Dir(mdPath)
	} else {
		workingDir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		cwd = filepath.Join(workingDir, filepath.Dir(mdPath))
	}

	return &RunContext{
		Registry:     NewRegistry(),
		ShellCmd:     cmd,
		Stdin:        stdin,
		Stdout:       pr,
		StdoutReader: bufio.NewReader(pr),
		Cwd:          cwd,
		waitDone:     waitDone,
	}, nil
}

func (ctx *RunContext) Cleanup() {
	if ctx.Stdin != nil {
		ctx.Stdin.Close()
	}
	if ctx.ShellCmd != nil && ctx.ShellCmd.Process != nil {
		ctx.ShellCmd.Process.Kill()
	}
	if ctx.waitDone != nil {
		<-ctx.waitDone
	}
	if ctx.Stdout != nil {
		ctx.Stdout.Close()
	}
	if ctx.Registry != nil {
		ctx.Registry.KillAll()
	}
}
