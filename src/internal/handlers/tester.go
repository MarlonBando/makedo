package handlers

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"makedo/internal/executor"
	"makedo/internal/nodes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// TestResult holds the result of a single directive test for reporting.
type TestResult struct {
	Passed    bool
	StartLine int
	Expected  string
	Actual    string
	Error     error
}

// lineNumber converts a byte offset to a 1-indexed line number
func lineNumber(source []byte, offset int) int {
	return bytes.Count(source[:offset], []byte{'\n'}) + 1
}

// testDirective tests a single directive against execution result
func testDirective(d *nodes.Directive, execResult *executor.Result, source []byte, startLine int, patterns map[*nodes.Directive]*regexp.Regexp) *TestResult {
	// Handle non-zero exit for completed commands
	if execResult.Status == executor.Completed && execResult.ExitCode != 0 {
		return &TestResult{
			Passed:    false,
			StartLine: startLine,
			Expected:  "command to succeed",
			Actual:    string(execResult.Output),
			Error:     fmt.Errorf("command exited with code %d", execResult.ExitCode),
		}
	}

	// If executor returned Ready, cmd directives already passed during execution
	if d.Kind == nodes.DirectiveCmd && execResult.Status == executor.Ready {
		return &TestResult{
			Passed:    true,
			StartLine: startLine,
			Expected:  "exit 0",
			Actual:    d.ContentString(source),
		}
	}

	// Use shared directive checking
	check := executor.CheckDirective(execResult.Output, d, source, patterns)
	return &TestResult{
		Passed:    check.Passed,
		StartLine: startLine,
		Expected:  check.Expected,
		Actual:    check.Actual,
		Error:     check.Err,
	}
}

func VerifyMarkdown(mdPath string) error {
	mdPath = strings.TrimSpace(mdPath)

	source, err := os.ReadFile(mdPath)
	if err != nil {
		return err
	}

	var allWarnings []string

	// Process registry for cleanup at document end
	registry := executor.NewRegistry()
	defer registry.KillAll()

	md := goldmark.New(
		goldmark.WithExtensions(nodes.NewMakeDoExtension()),
	)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var results []*TestResult
	testNum := 0

	// Walk AST and run tests
	walkErr := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if n.Kind() != nodes.KindMakeDoCodeBlock {
			return ast.WalkContinue, nil
		}

		block := n.(*nodes.MakeDoCodeBlock)
		directives := block.Directives()

		lines := block.Lines()
		startLine := lineNumber(source, lines.At(0).Start)

		code := block.Code(source)

		//patterns map is used to avoid recompiling the regex at every iteration in case
		//of a long running task like a server setup
		var patterns map[*nodes.Directive]*regexp.Regexp
		patterns, err = executor.PrecompileDirectives(directives, source)
		if err != nil {
			fmt.Printf("failed to precompile directives: %v\n", err)
			return ast.WalkContinue, nil
		}

		execResult := executor.Execute(string(code), directives, source, false)

		for _, warn := range execResult.Warnings {
			allWarnings = append(allWarnings, fmt.Sprintf("%s:%d: %v", mdPath, startLine, warn))
		}

		if execResult.Process != nil && execResult.Status != executor.Completed {
			// we add to registry all the process that are still running
			// so we can clean them up
			registry.Add(execResult.Process)
		}

		// No-directive shell blocks are setup blocks: run them but do not count as tests.
		if len(directives) == 0 {
			if execResult.Err != nil {
				return ast.WalkStop, fmt.Errorf("setup block at line %d failed: %w", startLine, execResult.Err)
			}
			if execResult.Status == executor.Completed && execResult.ExitCode != 0 {
				return ast.WalkStop, fmt.Errorf("setup block at line %d failed: command exited with code %d", startLine, execResult.ExitCode)
			}
			return ast.WalkContinue, nil
		}

		if execResult.Err != nil {
			testNum++
			fmt.Printf("test %d... failed\n", testNum)
			results = append(results, &TestResult{
				Passed:    false,
				StartLine: startLine,
				Expected:  "command to execute successfully",
				Actual:    string(code),
				Error:     execResult.Err,
			})
			return ast.WalkContinue, nil
		}

		// Handle stall: with directives = fail
		if execResult.Status == executor.Stalled {
			testNum++
			fmt.Printf("test %d... failed (stalled)\n", testNum)
			results = append(results, &TestResult{
				Passed:    false,
				StartLine: startLine,
				Expected:  "directives to pass before stall timeout",
				Actual:    "no output for 10s",
				Error:     fmt.Errorf("command stalled"),
			})
			return ast.WalkContinue, nil
		}

		// Test each directive for reporting
		for _, directive := range directives {
			result := testDirective(directive, execResult, source, startLine, patterns)
			if result == nil {
				continue
			}

			testNum++
			fmt.Printf("test %d... ", testNum)

			if result.Passed {
				fmt.Println("succeeded")
			} else {
				fmt.Println("failed")
			}

			results = append(results, result)
		}
		return ast.WalkContinue, nil
	})

	if walkErr != nil {
		return walkErr
	}

	// Print summary
	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}

	fmt.Println()
	fmt.Printf("=== Summary ===\n")
	fmt.Printf("%d/%d tests passed\n", passed, len(results))

	if len(allWarnings) > 0 {
		fmt.Println("\nWarnings Summary:")
		for _, w := range allWarnings {
			fmt.Printf("- %s\n", w)
		}
	}

	// Print failed tests details
	if passed < len(results) {
		fmt.Println()
		fmt.Println("Failed tests:")
		testNum = 0
		for _, r := range results {
			testNum++
			if r.Passed {
				continue
			}
			if r.Error != nil {
				fmt.Printf("  test %d (line %d): %v\n", testNum, r.StartLine, r.Error)
			} else {
				fmt.Printf("  test %d (line %d): pattern did not match\n", testNum, r.StartLine)
			}
			fmt.Printf("    expected: %s\n", r.Expected)
			fmt.Printf("    actual:   %s\n", r.Actual)
		}
		return fmt.Errorf("%d tests failed", len(results)-passed)
	}

	return nil
}
