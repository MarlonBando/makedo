package engine

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"makedo/internal/nodes"

	"github.com/yuin/goldmark/text"
)

func TestExecuteFastCommand(t *testing.T) {
	result := Execute(&RunContext{}, "echo hello", nil, nil, false)

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
	result := Execute(&RunContext{}, "exit 42", nil, nil, false)

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

	result := Execute(&RunContext{}, "echo hello world", []*nodes.Directive{directive}, source, false)

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

func TestExecuteStallsWithNoOutput(t *testing.T) {
	// Reduce stall timeout for test
	oldTimeout := StallTimeout
	StallTimeout = 500 * time.Millisecond
	defer func() { StallTimeout = oldTimeout }()

	result := Execute(&RunContext{}, "sleep 10", nil, nil, false)

	if result.Status != Stalled {
		t.Errorf("expected Stalled, got %v", result.Status)
	}
	if result.Process == nil {
		t.Error("expected process to be set for stalled command")
	}

	// Clean up
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
	result := Execute(&RunContext{}, "echo ready && sleep 10", []*nodes.Directive{directive}, source, false)
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

	start := time.Now()
	result := Execute(&RunContext{}, 
		"python3 -m http.server "+port+" --bind 127.0.0.1 2>&1",
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

func TestRegistryKillAll(t *testing.T) {
	registry := NewRegistry()

	// Use short stall timeout
	oldTimeout := StallTimeout
	StallTimeout = 200 * time.Millisecond
	result := Execute(&RunContext{}, "sleep 100", nil, nil, false)
	StallTimeout = oldTimeout

	if result.Process == nil {
		t.Fatal("expected stalled process")
	}

	registry.Add(result.Process)

	// Kill all should terminate the process
	registry.KillAll()

	// Process should be dead now
	done := make(chan struct{})
	go func() {
		result.Process.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Good - process was killed
	case <-time.After(1 * time.Second):
		t.Error("process was not killed by registry")
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

			result := CheckDirective(tt.output, directive, source, nil)
			if result.Passed != tt.wantPass {
				t.Errorf("CheckDirective().Passed = %v, want %v", result.Passed, tt.wantPass)
			}
		})
	}
}

func runShellCmd(cmd string) error {
	return Execute(&RunContext{}, cmd, nil, nil, false).Err
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

	result := CheckDirective(output, directive, source, patterns)
	if !result.Passed {
		t.Errorf("Expected output to contain 'Hello 2023-10-25', got failure")
	}
}
