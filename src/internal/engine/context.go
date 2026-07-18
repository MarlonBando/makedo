package engine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/creack/pty"
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
	// --- PTY INITIALIZATION EDGE CASES ---
	// Attaching a PTY makes bash think it's running in an interactive terminal.
	// This creates several major problems for automated directive matching:
	//
	// 1. Profiles: It evaluates ~/.bashrc and ~/.profile, polluting the environment.
	//    Fix: Pass --noprofile and --norc.
	//
	// 2. Prompts: It prints PS1 ("bash$ ") and PS2 ("> " for heredocs) which ruin matching.
	//    Fix: Explicitly clear PS1 and PS2 in the environment before starting.
	//
	// 3. Terminal ECHO: We must turn off ECHO so commands sent to stdin aren't printed
	//    to stdout. Cross-platform ioctl syscalls (TCGETS vs TIOCGETA) are fragile across
	//    Linux/macOS, so we natively run `stty -echo` inside the shell instead.
	//
	// 4. Readline Race Condition: Interactive bash uses the GNU readline library, which
	//    automatically forces ECHO back ON after initialization, overriding our `stty`!
	//    Fix: Pass --noediting to completely disable readline and make `stty` stick.
	cmd := exec.Command(shell, "--noprofile", "--norc", "--noediting")
	cmd.Env = append(os.Environ(), "PS1=", "PS2=")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	stdoutReader := bufio.NewReader(ptmx)

	// Cleanly disable terminal ECHO using the shell itself, and flush the startup buffer.
	fmt.Fprintln(ptmx, "stty -echo 2>/dev/null; echo MAKEDO_PTY_READY")
	for {
		line, _ := stdoutReader.ReadString('\n')
		if strings.Contains(line, "MAKEDO_PTY_READY") && !strings.Contains(line, "stty") {
			break
		}
	}

	// We need to wait for the command to finish so we can close waitDone
	waitDone := make(chan struct{})
	go func() {
		_ = cmd.Wait()
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
		Stdin:        ptmx,
		Stdout:       ptmx,
		StdoutReader: stdoutReader,
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
