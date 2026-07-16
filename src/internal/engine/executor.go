package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
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

func Execute(ctx *RunContext, code string, directives []*nodes.Directive, source []byte, stream bool) *Result {
	compiledPatterns, err := PrecompileDirectives(directives, source)
	if err != nil {
		return &Result{Err: err, ExitCode: -1}
	}

	if hasBackgroundOperator(code) {
		return executeBackground(ctx, code, directives, source, stream, compiledPatterns)
	}

	return executeSync(ctx, code, directives, source, stream, compiledPatterns)
}

func executeSync(ctx *RunContext, code string, directives []*nodes.Directive, source []byte, stream bool, compiledPatterns map[*nodes.Directive]*regexp.Regexp) *Result {
	uuidVal := time.Now().UnixNano()
	cwdFile := fmt.Sprintf("/tmp/makedo_cwd_%d", uuidVal)
	envFile := fmt.Sprintf("/tmp/makedo_env_%d", uuidVal)
	markerPrefix := fmt.Sprintf("MAKEDO_DONE_%d:", uuidVal)

	// The empty echo "" is CRITICAL: If the user's command does not output a trailing newline
	// (e.g. `printf` or `curl -w`), it will concatenate with our marker on the same line,
	// preventing strings.HasPrefix from detecting the marker and causing a deadlock.
	//
	// Capturing EXIT_CODE=$? immediately after the user code ensures we track the exit status
	// of the exact last command the user wrote, mimicking standard bash behavior.
	script := fmt.Sprintf(`%s
EXIT_CODE=$?
pwd > %s
export -p > %s
echo ""
echo "%s $EXIT_CODE"
`, code, cwdFile, envFile, markerPrefix)

	_, err := fmt.Fprint(ctx.Stdin, script)
	if err != nil {
		return &Result{Err: err, ExitCode: -1}
	}

	var buf bytes.Buffer
	for {
		line, err := ctx.StdoutReader.ReadString('\n')
		if err != nil {
			return &Result{Err: fmt.Errorf("shell environment crashed: %w", err), ExitCode: -1}
		}

		if strings.HasPrefix(line, markerPrefix) {
			exitCodeStr := strings.TrimSpace(strings.TrimPrefix(line, markerPrefix))
			exitCode := 0
			fmt.Sscanf(exitCodeStr, "%d", &exitCode)

			// Update state
			if cwdBytes, err := os.ReadFile(cwdFile); err == nil {
				ctx.Cwd = strings.TrimSpace(string(cwdBytes))
			}
			os.Remove(cwdFile)

			// Clean up previous env file
			if ctx.EnvFile != "" {
				os.Remove(ctx.EnvFile)
			}
			ctx.EnvFile = envFile

			return &Result{
				Status:   Completed,
				Output:   buf.Bytes(),
				ExitCode: exitCode,
			}
		}

		// We forward the output of the block to the shell
		// in which makedo is running (if it's not the internal marker)
		if stream {
			_, _ = os.Stdout.WriteString(line)
		}

		buf.WriteString(line)
	}
}

func executeBackground(ctx *RunContext, code string, directives []*nodes.Directive, source []byte, stream bool, compiledPatterns map[*nodes.Directive]*regexp.Regexp) *Result {
	uuidVal := time.Now().UnixNano()
	cwdFile := fmt.Sprintf("/tmp/makedo_cwd_%d", uuidVal)
	envFile := fmt.Sprintf("/tmp/makedo_env_%d", uuidVal)
	markerPrefix := fmt.Sprintf("MAKEDO_DONE_%d:", uuidVal)

	// Dump state from persistent shell.
	// The empty echo "" prevents concatenation issues if the previous command omitted a newline.
	script := fmt.Sprintf(`pwd > %s
export -p > %s
echo ""
echo '%s 0'
`, cwdFile, envFile, markerPrefix)
	_, err := fmt.Fprint(ctx.Stdin, script)

	for {
		line, err := ctx.StdoutReader.ReadString('\n')
		if err != nil {
			return &Result{Err: fmt.Errorf("shell environment crashed during state handoff: %w", err), ExitCode: -1}
		}
		if strings.HasPrefix(line, markerPrefix) {
			break
		}
	}

	// We append `\nwait` so that `bash` blocks waiting for the `&` background processes to finish.
	// This keeps the output pipe open so `monitor()` can continuously poll the output and directives,
	// rather than `bash` exiting instantly and returning EOF.
	bgCode := fmt.Sprintf(`cd "$(cat %s)"
source %s
rm -f %s %s
%s
wait`, cwdFile, envFile, cwdFile, envFile, code)
	shell, flag := getShell()
	cmd := exec.Command(shell, flag, bgCode)

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

	result := monitor(cmd, output, done, directives, source, stream, compiledPatterns, ctx)
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

	// If it's a background block with no directives, we should not block waiting for EOF.
	// We return immediately, just like bash does when you append `&`.
	if !hasDirectives {
		close(done)
		return &Result{
			Status:   Ready,
			Output:   buf.Bytes(),
			Process:  cmd.Process,
			ExitCode: -1,
		}
	}

	needsCmdCheck := hasCmdDirective(directives)
	allowEarlyExit := canEarlyExit(directives)

	// Immediate initial check of ALL directives (before any output)
	// This handles cases where directives pass before command produces output
	if allowEarlyExit && checkAllDirectives(buf.Bytes(), directives, source, compiledPatterns, ctx) {
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
// Draining is essential because if we stop reading, a background shell
// process could eventually block entirely on a full stdout buffer and freeze.
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
//
// TODO: This uses basic string manipulation and will yield false positives
// in edge cases like stderr redirects (`>& 2`), subshells (`$(cmd &)`),
// and heredocs. It is worth considering replacing this with a full bash AST
// parser (like mvdan.cc/sh) in the future for perfect accuracy.
func hasBackgroundOp(b string, strRanges []byteRange) bool {
	parens := 0
	for i := 0; i < len(b); i++ {
		if inAnyRange(i, strRanges) {
			continue
		}

		if b[i] == '(' {
			parens++
			continue
		}
		if b[i] == ')' {
			parens--
			continue
		}

		if b[i] != '&' {
			continue
		}

		// Skip if we are inside a subshell like $(cmd &)
		if parens > 0 {
			continue
		}

		// Skip >& (stderr redirection)
		if i > 0 && b[i-1] == '>' {
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

func getShell() (shell string, flag string) {
	if sh := os.Getenv("SHELL"); sh != "" {
		return sh, "-c"
	}

	// TODO: Support Windows environments in the future
	if runtime.GOOS == "windows" {
		if comspec := os.Getenv("COMSPEC"); comspec != "" {
			return comspec, "/c"
		}
		return "cmd.exe", "/c"
	}

	// Fallback for Unix/macOS/Linux systems if $SHELL is unset
	return "/bin/sh", "-c"
}
