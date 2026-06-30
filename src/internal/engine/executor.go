package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"makedo/internal/nodes"
)

const (
	TickerInterval = 100 * time.Millisecond
	bufSize        = 4096
)

// Execute runs a command with goroutine-based output monitoring.
// Returns when: command exits, all directives pass.
func Execute(ctx *RunContext, code string, directives []*nodes.Directive, source []byte, stream bool) *Result {
	// before every shell execution we load the makedo env file
	// in this way the user can share varibles
	// across different shell block
	code = fmt.Sprintf("source %q\n%s", ctx.MkEnvFile, code)

	compiledPatterns, err := PrecompileDirectives(directives, source)
	if err != nil {
		return &Result{Err: err, ExitCode: -1}
	}

	var warnings []error
	if hasBackgroundOperator(code) {
		warnings = append(warnings, fmt.Errorf("detected & in block! makedo manages background processes automatically. Split long-running commands into their own block and use a directive to signal readiness"))
	}

	cmd := exec.Command(os.Getenv("SHELL"), "-c", code)

	// Set up process group so we can kill the entire tree
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return &Result{Err: err, ExitCode: -1, Warnings: warnings}
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return &Result{Err: err, ExitCode: -1, Warnings: warnings}
	}

	if err := cmd.Start(); err != nil {
		return &Result{Err: err, ExitCode: -1, Warnings: warnings}
	}

	// Channel for output chunks, closed when both readers exit
	output := make(chan []byte, 16)
	done := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)
	// TODO: stdout and stderr are multiplexed into the same output channel.
	// This means 'out' and 'outr' directives check the combined stream.
	// Revisit this if we want strict separation (e.g., implementing an 'err' directive for stderr only).
	go pumpOutput(stdout, output, done, &wg)
	go pumpOutput(stderr, output, done, &wg)
	go func() {
		wg.Wait()
		close(output)
	}()

	result := monitor(cmd, output, done, directives, source, stream, compiledPatterns, ctx)
	result.Warnings = append(result.Warnings, warnings...)

	// If process still running, it stays alive for registry to kill later
	return result
}

// False if all directives are negated
// This is important because if we check that a text is different from "Server ready" for example
// as soon as the first stream arrive we don't see "Server ready" the test passes but it should wait
func canEarlyExit(directives []*nodes.Directive) bool {
	hasNegativeOut := false
	hasPositiveOut := false

	for _, d := range directives {
		if d.Kind == nodes.DirectiveOut || d.Kind == nodes.DirectiveOutRegex {
			hasNegativeOut = hasNegativeOut || d.Negated
			hasPositiveOut = hasPositiveOut || !d.Negated
		}
	}

	return !hasNegativeOut || hasPositiveOut
}

// TODO: Improve efficiency. We run all the directives when checking. Maybe some directives already passed
// so technically we shouldn't be checking them. But this is for later.
func monitor(cmd *exec.Cmd, output <-chan []byte, done chan struct{}, directives []*nodes.Directive, source []byte, stream bool, compiledPatterns map[*nodes.Directive]*regexp.Regexp, ctx *RunContext) *Result {
	var buf bytes.Buffer
	ticker := time.NewTicker(TickerInterval)
	defer ticker.Stop()

	hasDirectives := len(directives) > 0
	needsCmdCheck := hasCmdDirective(directives)
	allowEarlyExit := canEarlyExit(directives)

	// Immediate initial check of ALL directives (before any output)
	// This handles cases where directives pass before command produces output
	if hasDirectives && allowEarlyExit && checkAllDirectives(buf.Bytes(), directives, source, compiledPatterns, ctx) {
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

			if stream {
				_, _ = os.Stdout.Write(chunk)
			}

			// Fast-track: check ONLY fast directives (out, outr, pwd) on output
			// cmd directives are checked on ticker to avoid spawning too many processes
			if hasDirectives && !needsCmdCheck && allowEarlyExit && checkFastDirectives(buf.Bytes(), directives, source, compiledPatterns, ctx) {
				close(done)
				return &Result{
					Status:   Ready,
					Output:   buf.Bytes(),
					Process:  cmd.Process,
					ExitCode: -1,
				}
			}

		case <-ticker.C:
			// Full directive check (includes cmd) on ticker
			// Runs synchronously - only one cmd process at a time
			if hasDirectives && allowEarlyExit && checkAllDirectives(buf.Bytes(), directives, source, compiledPatterns, ctx) {
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

type byteRange struct{ start, end int }

// stringRanges scans a single line and returns the byte ranges occupied
// by quoted strings. Single-quoted strings have no escape sequences.
// Double-quoted strings honour backslash escapes (e.g. \").
func stringRanges(b string, ranges []byteRange) []byteRange {
	ranges = ranges[:0]
	i := 0
	for i < len(b) {
		switch b[i] {
		case '\'':
			start := i
			i++
			for i < len(b) && b[i] != '\'' {
				i++
			}
			if i < len(b) {
				i++ // consume closing '
			}
			ranges = append(ranges, byteRange{start, i})

		case '"':
			start := i
			i++
			for i < len(b) && b[i] != '"' {
				if b[i] == '\\' && i+1 < len(b) {
					i++ // skip the escaped character
				}
				i++
			}
			if i < len(b) {
				i++ // consume closing "
			}
			ranges = append(ranges, byteRange{start, i})

		default:
			i++
		}
	}
	return ranges
}

// inAnyRange reports whether pos falls inside any of the given ranges.
func inAnyRange(pos int, ranges []byteRange) bool {
	for _, r := range ranges {
		if pos >= r.start && pos < r.end {
			return true
		}
	}
	return false
}

// stripComment returns the slice of b up to the first # that falls
// outside a quoted string, trimming trailing whitespace.
// strRanges must have been computed on b before calling.
func stripComment(b string, strRanges []byteRange) string {
	for i := 0; i < len(b); i++ {
		if b[i] == '#' && !inAnyRange(i, strRanges) {
			return strings.TrimRight(b[:i], " \t")
		}
	}
	return b
}

// hasBackgroundOp scans b for a lone & (not part of &&) that is outside
// any quoted string and is either followed by a space/tab or sits at
// the end of the slice — the unambiguous shell background operator.
// strRanges must have been computed on the original line before comment
// stripping; they remain valid because stripComment only shortens from
// the right, never before any string range.
func hasBackgroundOp(b string, strRanges []byteRange) bool {
	for i := 0; i < len(b); i++ {
		if b[i] != '&' {
			continue
		}
		if inAnyRange(i, strRanges) {
			continue
		}
		// && → logical AND, skip both characters.
		if i+1 < len(b) && b[i+1] == '&' {
			i++
			continue
		}
		// & followed by space/tab or sitting at end of line.
		if i+1 == len(b) || b[i+1] == ' ' || b[i+1] == '\t' {
			return true
		}
	}
	return false
}

// HasBackgroundOp walks each line of a shell code block and returns
// the first line that contains a background operator outside of any
// quoted string or comment.
func hasBackgroundOperator(code string) bool {
	var strRanges []byteRange
	for len(code) > 0 {
		var line string
		idx := strings.IndexByte(code, '\n')
		if idx >= 0 {
			line = code[:idx]
			code = code[idx+1:]
		} else {
			line = code
			code = ""
		}

		trimmed := strings.TrimSpace(line)

		// Skip blank lines and full-line comments.
		if len(trimmed) == 0 || trimmed[0] == '#' {
			continue
		}

		strRanges = stringRanges(trimmed, strRanges)
		effective := stripComment(trimmed, strRanges)

		if hasBackgroundOp(effective, strRanges) {
			return true
		}
	}

	return false
}
