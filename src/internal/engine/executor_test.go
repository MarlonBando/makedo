package engine

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"makedo/internal/nodes"

	"github.com/yuin/goldmark/text"
)

func TestExecuteFastCommand(t *testing.T) {
	ctx, _ := NewRunContext("")
	defer ctx.Cleanup()
	result := Execute(ctx, "echo hello", nil, nil, false)

	if result.Status != Completed {
		t.Errorf("expected Completed, got %v", result.Status)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if !strings.Contains(string(result.Output), "hello") {
		t.Errorf("expected output to contain 'hello', got %q", result.Output)
	}
}

func TestExecuteCommandWithNonZeroExit(t *testing.T) {
	ctx, _ := NewRunContext("")
	defer ctx.Cleanup()
	result := Execute(ctx, "sh -c 'exit 42'", nil, nil, false)

	if result.Status != Completed {
		t.Errorf("expected Completed, got %v", result.Status)
	}
	if result.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", result.ExitCode)
	}
}

func TestExecuteWithOutDirective(t *testing.T) {
	source := []byte("<!-- out hello -->")
	directive := &nodes.Directive{
		Kind:    nodes.DirectiveOut,
		Content: text.NewSegment(9, 14), // "hello"
	}

	ctx, _ := NewRunContext("")
	defer ctx.Cleanup()
	result := Execute(ctx, "echo hello world", []*nodes.Directive{directive}, source, false)

	// Can be either Completed (if command exits before we check) or Ready (if directive passes first)
	if result.Status != Completed && result.Status != Ready {
		t.Errorf("expected Completed or Ready, got %v", result.Status)
	}
	if !strings.Contains(string(result.Output), "hello") {
		t.Errorf("expected output to contain 'hello', got %q", result.Output)
	}

	// Clean up if process still running
	if result.Process != nil {
		result.Process.Kill()
		result.Process.Wait()
	}
}

func TestExecuteLongRunningWithDirective(t *testing.T) {
	source := []byte("<!-- out ready -->")
	directive := &nodes.Directive{
		Kind:    nodes.DirectiveOut,
		Content: text.NewSegment(9, 14), // "ready"
	}

	start := time.Now()
	// This command outputs "ready" then sleeps - but it completes quickly
	// because shell sees echo finish before sleep starts in pipeline

	ctx, _ := NewRunContext("")
	defer ctx.Cleanup()
	result := Execute(ctx, "echo ready && sleep 10 &", []*nodes.Directive{directive}, source, false)
	elapsed := time.Since(start)

	// Should complete much faster than 10s (directive matches immediately)
	if elapsed > 2*time.Second {
		t.Errorf("expected quick return when directive passes, took %v", elapsed)
	}

	if !strings.Contains(string(result.Output), "ready") {
		t.Errorf("expected output to contain 'ready', got %q", result.Output)
	}

	// Clean up if process still running
	if result.Process != nil {
		result.Process.Kill()
		result.Process.Wait()
	}
}

func TestExecuteServerWithCurlDirective(t *testing.T) {
	// Skip if python3 not available
	if err := runShellCmd("which python3"); err != nil {
		t.Skip("python3 not available")
	}

	// Use time-based port to avoid conflicts between test runs
	port := fmt.Sprintf("%d", 19000+time.Now().UnixNano()%1000)

	source := []byte("<!-- cmd curl -sf http://127.0.0.1:" + port + "/ >/dev/null -->")
	directive := &nodes.Directive{
		Kind:    nodes.DirectiveCmd,
		Content: text.NewSegment(9, len(source)-4), // the curl command
	}

	ctx, _ := NewRunContext("")
	defer ctx.Cleanup()
	start := time.Now()
	result := Execute(ctx,
		"python3 -m http.server "+port+" --bind 127.0.0.1 2>&1 &",
		[]*nodes.Directive{directive},
		source,
		false,
	)
	elapsed := time.Since(start)

	// Should return Ready (server running, curl succeeded)
	if result.Status != Ready {
		t.Errorf("expected Ready, got %v (elapsed: %v, output: %s)", result.Status, elapsed, result.Output)
	}

	// Should complete in a reasonable time
	if elapsed > 10*time.Second {
		t.Errorf("expected completion within 10s, took %v", elapsed)
	}

	// Process should still be running
	if result.Process == nil {
		t.Error("expected process to be set")
	}

	// Clean up
	if result.Process != nil {
		result.Process.Kill()
		result.Process.Wait()
	}
}

func TestCheckDirectives(t *testing.T) {
	tests := []struct {
		name     string
		output   []byte
		kind     nodes.DirectiveKind
		content  string
		wantPass bool
	}{
		{
			name:     "out matches substring",
			output:   []byte("hello world"),
			kind:     nodes.DirectiveOut,
			content:  "world",
			wantPass: true,
		},
		{
			name:     "out does not match",
			output:   []byte("hello world"),
			kind:     nodes.DirectiveOut,
			content:  "foo",
			wantPass: false,
		},
		{
			name:     "outr matches regex",
			output:   []byte("version 2.34.1"),
			kind:     nodes.DirectiveOutRegex,
			content:  `version \d+\.\d+`,
			wantPass: true,
		},
		{
			name:     "outr does not match",
			output:   []byte("version 2.34.1"),
			kind:     nodes.DirectiveOutRegex,
			content:  `^version 1\.`,
			wantPass: false,
		},
		{
			name:     "checkpath matches current dir",
			output:   []byte(""),
			kind:     nodes.DirectiveCheckpath,
			content:  ".",
			wantPass: true,
		},
		{
			name:     "checkpath does not match non-existent file",
			output:   []byte(""),
			kind:     nodes.DirectiveCheckpath,
			content:  "non_existent_file_12345",
			wantPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.content)
			directive := &nodes.Directive{
				Kind:    tt.kind,
				Content: text.NewSegment(0, len(tt.content)),
			}

			result := CheckDirective(tt.output, directive, source, nil, &RunContext{})
			if result.Passed != tt.wantPass {
				t.Errorf("CheckDirective().Passed = %v, want %v", result.Passed, tt.wantPass)
			}
		})
	}
}

func runShellCmd(cmd string) error {
	ctx, _ := NewRunContext("")
	defer ctx.Cleanup()
	return Execute(ctx, cmd, nil, nil, false).Err
}

func TestCheckDirectiveTypeExpansion(t *testing.T) {
	source := []byte("<!-- out Hello ${{date}} -->")
	directive, _ := nodes.ParseDirective(source, 0)
	output := []byte("Hello 2023-10-25\n")

	// Needs to precompile
	patterns, err := PrecompileDirectives([]*nodes.Directive{directive}, source)
	if err != nil {
		t.Fatalf("Precompile failed: %v", err)
	}

	result := CheckDirective(output, directive, source, patterns, &RunContext{})
	if !result.Passed {
		t.Errorf("Expected output to contain 'Hello 2023-10-25', got failure")
	}
}

func TestCheckDirectiveCmdWithEnv(t *testing.T) {
	ctx, err := NewRunContext("")
	if err != nil {
		t.Fatalf("failed to create run context: %v", err)
	}
	defer ctx.Cleanup()

	// Append environment variable definition to the environment file
	envData := "export TEST_VAR=my_secret_val\n"
	envFile := filepath.Join(os.TempDir(), "test_env_file_cmd")
	os.WriteFile(envFile, []byte(envData), 0600)
	ctx.EnvFile = envFile
	defer os.Remove(envFile)

	source := []byte(`<!-- cmd [ "$TEST_VAR" = "my_secret_val" ] -->`)
	directive, ok := nodes.ParseDirective(source, 0)
	if !ok {
		t.Fatalf("failed to parse directive")
	}

	result := CheckDirective(nil, directive, source, nil, ctx)
	if !result.Passed {
		t.Errorf("expected cmd directive to pass with env sourced, got error: %v", result.Err)
	}
}

func TestCleanOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No ANSI",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Carriage Returns",
			input:    "Line1\r\nLine2\r\n",
			expected: "Line1\nLine2\n",
		},
		{
			name:     "Simple ANSI Color",
			input:    "\x1b[32mColorful\x1b[0m",
			expected: "Colorful",
		},
		{
			name:     "Complex ANSI Codes",
			input:    "\x1b[1;31;42mError\x1b[0m\r\n",
			expected: "Error\n",
		},
		{
			name:     "Multiple ANSI Codes",
			input:    "\x1b[31mRed\x1b[0m \x1b[32mGreen\x1b[0m",
			expected: "Red Green",
		},
		{
			name:     "ANSI Sequence Not Recognized (No Bracket)",
			input:    "\x1bA text",
			expected: "\x1bA text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := CleanOutput([]byte(tt.input))
			if string(actual) != tt.expected {
				t.Errorf("CleanOutput() = %q, expected %q", actual, tt.expected)
			}
		})
	}
}

func TestCleanWriter(t *testing.T) {
	tests := []struct {
		name     string
		chunks   []string
		expected string
	}{
		{
			name:     "All in one chunk",
			chunks:   []string{"\x1b[32mGreen\x1b[0m\r\n"},
			expected: "Green\n",
		},
		{
			name:     "Split after ESC",
			chunks:   []string{"Hello \x1b", "[32mWorld\x1b[0m"},
			expected: "Hello World",
		},
		{
			name:     "Split inside ESC bracket",
			chunks:   []string{"Hello \x1b[", "32mWorld\x1b[0m"},
			expected: "Hello World",
		},
		{
			name:     "Split inside parameters",
			chunks:   []string{"Hello \x1b[3", "2mWorld\x1b[0m"},
			expected: "Hello World",
		},
		{
			name:     "Split just before terminator",
			chunks:   []string{"Hello \x1b[32", "mWorld\x1b[0m"},
			expected: "Hello World",
		},
		{
			name:     "Split carriage returns",
			chunks:   []string{"Line1\r", "\nLine2\r", "\n"},
			expected: "Line1\nLine2\n",
		},
		{
			name:     "Multiple tiny chunks",
			chunks:   []string{"\x1b", "[", "3", "1", "m", "R", "e", "d", "\x1b", "[", "0", "m"},
			expected: "Red",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			cleaner := NewCleanWriter(&out)

			for _, chunk := range tt.chunks {
				n, err := cleaner.Write([]byte(chunk))
				if err != nil {
					t.Fatalf("Write() error = %v", err)
				}
				if n != len(chunk) {
					t.Fatalf("Write() n = %d, expected %d", n, len(chunk))
				}
			}

			if out.String() != tt.expected {
				t.Errorf("CleanWriter output = %q, expected %q", out.String(), tt.expected)
			}
		})
	}
}
