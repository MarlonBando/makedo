package handlers

import (
	"bytes"
	"fmt"
	"os"
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
func testDirective(d *nodes.Directive, execResult *executor.Result, source []byte, startLine int) *TestResult {
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
	check := executor.CheckDirective(execResult.Output, d, source)
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
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if n.Kind() != nodes.KindMakeDoCodeBlock {
			return ast.WalkContinue, nil
		}

		block := n.(*nodes.MakeDoCodeBlock)
		directives := block.Directives()

		// Skip blocks without directives
		if len(directives) == 0 {
			return ast.WalkContinue, nil
		}

		lines := block.Lines()
		startLine := lineNumber(source, lines.At(0).Start)

		// Execute block using executor
		code := block.Code(source)
		execResult := executor.Execute(string(code), directives, source, false)

		// Register process for cleanup if still running
		if execResult.Process != nil && execResult.Status != executor.Completed {
			registry.Add(execResult.Process)
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
			result := testDirective(directive, execResult, source, startLine)
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
	}

	return nil
}
