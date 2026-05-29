package handlers

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"makedo/internal/engine"
	"makedo/internal/nodes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// lineNumber converts a byte offset to a 1-indexed line number
func lineNumber(source []byte, offset int) int {
	return bytes.Count(source[:offset], []byte{'\n'}) + 1
}

func VerifyMarkdown(mdPath string, ctx *engine.RunContext) error {
	mdPath = strings.TrimSpace(mdPath)

	source, err := os.ReadFile(mdPath)
	if err != nil {
		return err
	}

	var allWarnings []string

	md := goldmark.New(
		goldmark.WithExtensions(nodes.NewMakeDoExtension()),
	)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var results []*engine.TestResult
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

		code := string(block.Code(source))

		outcome := engine.EvaluateBlock(code, directives, source, startLine, ctx)

		for _, warn := range outcome.ExecResult.Warnings {
			allWarnings = append(allWarnings, fmt.Sprintf("%s:%d: %v", mdPath, startLine, warn))
		}

		// No-directive shell blocks are setup blocks: run them but do not count as tests.
		if len(directives) == 0 {
			if !outcome.Passed {
				return ast.WalkStop, fmt.Errorf("setup block at line %d failed: %w", startLine, outcome.FailReason)
			}
			return ast.WalkContinue, nil
		}

		// Handle stall, execution err, or command fail that happened before directives
		// EvaluateBlock already catches these, but we need to print test status
		if outcome.ExecResult.Err != nil || (outcome.ExecResult.Status == engine.Stalled && len(directives) > 0) || (outcome.ExecResult.Status == engine.Completed && outcome.ExecResult.ExitCode != 0) {
			// Actually, if it failed on the block level rather than a specific directive, we might not have a TestResult.
			if len(outcome.TestResults) == 0 {
				testNum++
				fmt.Printf("test %d... failed\n", testNum)
				results = append(results, &engine.TestResult{
					Passed:    false,
					StartLine: startLine,
					Expected:  "command to execute successfully without stalling",
					Actual:    fmt.Sprintf("failed with: %v", outcome.FailReason),
					Error:     outcome.FailReason,
				})
				return ast.WalkContinue, nil
			}
		}

		// Print reporting for individual directives
		for _, testRes := range outcome.TestResults {
			testNum++
			fmt.Printf("test %d... ", testNum)

			if testRes.Passed {
				fmt.Println("succeeded")
			} else {
				fmt.Println("failed")
			}

			results = append(results, testRes)
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
