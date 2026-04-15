package executor

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"syscall"
	"time"

	"makedo/internal/nodes"
)

const (
	TickerInterval = 100 * time.Millisecond
	bufSize        = 4096
)

// StallTimeout is the duration of no output before a command is considered stalled.
// Variable to allow testing with shorter timeouts.
var StallTimeout = 10 * time.Second

// Execute runs a command with goroutine-based output monitoring.
// Returns when: command exits, all directives pass, or output stalls.
func Execute(code string, directives []*nodes.Directive, source []byte, stream bool) *Result {
	compiledPatterns, err := PrecompileDirectives(directives, source)
	if err != nil {
		return &Result{Err: err, ExitCode: -1}
	}

	cmd := exec.Command(os.Getenv("SHELL"), "-c", code)

	// Set up process group so we can kill the entire tree
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return &Result{Err: err, ExitCode: -1}
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return &Result{Err: err, ExitCode: -1}
	}

	if err := cmd.Start(); err != nil {
		return &Result{Err: err, ExitCode: -1}
	}

	// Channel for output chunks, closed when both readers exit
	output := make(chan []byte, 16)
	done := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)
	go pumpOutput(stdout, output, done, &wg)
	go pumpOutput(stderr, output, done, &wg)
	go func() {
		wg.Wait()
		close(output)
	}()

	result := monitor(cmd, output, done, directives, source, stream, compiledPatterns)

	// If process still running, it stays alive for registry to kill later
	return result
}

// TODO: Improve efficiency. We run all the directives when checking. Maybe some directives already passed
// so technically we shouldn't be checking them. But this is for later.
func monitor(cmd *exec.Cmd, output <-chan []byte, done chan struct{}, directives []*nodes.Directive, source []byte, stream bool, compiledPatterns map[*nodes.Directive]*regexp.Regexp) *Result {
	var buf bytes.Buffer
	lastActivity := time.Now()
	ticker := time.NewTicker(TickerInterval)
	defer ticker.Stop()

	hasDirectives := len(directives) > 0
	needsCmdCheck := hasCmdDirective(directives)

	// Immediate initial check of ALL directives (before any output)
	// This handles cases where directives pass before command produces output
	if hasDirectives && checkAllDirectives(buf.Bytes(), directives, source, compiledPatterns) {
		close(done)
		return &Result{
			Status:   Ready,
			Output:   buf.Bytes(),
			Process:  cmd.Process,
			ExitCode: -1,
		}
	}

	for {
		select {
		case chunk, ok := <-output:
			if !ok {
				// Channel closed - both readers done, command exited
				_ = cmd.Wait()
				exitCode := 0
				if cmd.ProcessState != nil && !cmd.ProcessState.Success() {
					exitCode = cmd.ProcessState.ExitCode()
				}
				return &Result{
					Status:   Completed,
					Output:   buf.Bytes(),
					ExitCode: exitCode,
				}
			}

			buf.Write(chunk)
			lastActivity = time.Now()

			if stream {
				_, _ = os.Stdout.Write(chunk)
			}

			// Fast-track: check ONLY fast directives (out, outr, pwd) on output
			// cmd directives are checked on ticker to avoid spawning too many processes
			if hasDirectives && !needsCmdCheck && checkFastDirectives(buf.Bytes(), directives, source, compiledPatterns) {
				close(done)
				return &Result{
					Status:   Ready,
					Output:   buf.Bytes(),
					Process:  cmd.Process,
					ExitCode: -1,
				}
			}

		case <-ticker.C:
			// Stall check
			if time.Since(lastActivity) > StallTimeout {
				close(done)
				// Stalled + no directives = pass (assumed ready)
				// Stalled + directives not passed = fail (but we return Stalled, caller decides)
				return &Result{
					Status:   Stalled,
					Output:   buf.Bytes(),
					Process:  cmd.Process,
					ExitCode: -1,
				}
			}

			// Full directive check (includes cmd) on ticker
			// Runs synchronously - only one cmd process at a time
			if hasDirectives && checkAllDirectives(buf.Bytes(), directives, source, compiledPatterns) {
				close(done)
				return &Result{
					Status:   Ready,
					Output:   buf.Bytes(),
					Process:  cmd.Process,
					ExitCode: -1,
				}
			}
		}
	}
}

// pumpOutput reads from r and sends to ch until EOF or done signal.
// When done is closed, continues reading (drain mode) but discards data.
func pumpOutput(r io.Reader, ch chan<- []byte, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	buf := make([]byte, bufSize)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			select {
			case <-done:
				// Drain mode: keep reading but discard
			default:
				// Normal mode: send chunk
				// Here we allocate again memory to avoid synchronizing the buf
				// because since we send the pointer to buf we need to be sure that the reader
				// will not write in the buf while it's processed.
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				ch <- chunk
			}
		}
		if err != nil {
			return
		}
	}
}
