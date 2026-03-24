package handlers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"makedo/internal/nodes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type TestResult struct {
	Passed    bool
	StartLine int
	Expected  string
	Actual    string
	Error     error
}

// lineNumber converts a byte offset to a 1-indexed line number
func lineNumber(source []byte, offset int) int {
	//TODO: what if we have a \n in a comment? We shouldn't count that???
	return bytes.Count(source[:offset], []byte{'\n'}) + 1
}

func out(block *nodes.MakeDoCodeBlock, source []byte) *TestResult {
	directive := block.GetDirective("out", source)
	if directive == nil {
		return nil
	}

	// TODO: what if multiple out directives?

	lines := block.Lines()
	startOffset := lines.At(0).Start
	startLine := lineNumber(source, startOffset)

	expected := directive.ContentString(source)
	code := block.Code(source)

	// Execute the command
	cmd := exec.Command(os.Getenv("SHELL"), "-c", string(code))
	output, err := cmd.Output()

	// If execution error, fail immediately without checking output
	if err != nil {
		return &TestResult{
			Passed:    false,
			StartLine: startLine,
			Expected:  expected,
			Actual:    string(output),
			Error:     err,
		}
	}

	// TODO: how to handle new lines in stdout???
	actual := strings.TrimRight(string(output), "\n")

	//NOTE: It checks if it's contained by default
	re, err := regexp.Compile(expected)
	if err != nil {
		return &TestResult{
			Passed:    false,
			StartLine: startLine,
			Expected:  expected,
			Actual:    actual,
			Error:     fmt.Errorf("Your regex expression '%s' is invalid: %w", expected, err),
		}
	}

	if !re.MatchString(actual) {
		return &TestResult{
			Passed:    false,
			StartLine: startLine,
			Expected:  expected,
			Actual:    actual,
			Error:     nil,
		}
	}

	return &TestResult{
		Passed:    true,
		StartLine: startLine,
		Expected:  expected,
		Actual:    actual,
		Error:     nil,
	}
}

func VerifyMarkdown(mdPath string) error {
	mdPath = strings.TrimSpace(mdPath)

	source, err := os.ReadFile(mdPath)
	if err != nil {
		return err
	}

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
		result := out(block, source)
		if result == nil {
			// No "out" directive, skip
			return ast.WalkContinue, nil
		}

		testNum++
		fmt.Printf("test %d... ", testNum)

		if result.Passed {
			fmt.Println("succeeded")
		} else {
			fmt.Println("failed")
		}

		results = append(results, result)
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
