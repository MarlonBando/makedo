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

// blockResult holds the result of executing a code block
type blockResult struct {
	output []byte
	err    error
}

// lineNumber converts a byte offset to a 1-indexed line number
func lineNumber(source []byte, offset int) int {
	return bytes.Count(source[:offset], []byte{'\n'}) + 1
}

func execBlock(block *nodes.MakeDoCodeBlock, source []byte) blockResult {
	code := block.Code(source)
	cmd := exec.Command(os.Getenv("SHELL"), "-c", string(code))
	output, err := cmd.CombinedOutput()
	return blockResult{output: output, err: err}
}

func out(block *nodes.MakeDoCodeBlock, source []byte, br blockResult, startLine int) *TestResult {
	directive := block.GetDirective(nodes.DirectiveOut)
	if directive == nil {
		return nil
	}

	expected := strings.TrimSpace(directive.ContentString(source))

	if br.err != nil {
		return &TestResult{
			Passed:    false,
			StartLine: startLine,
			Expected:  expected,
			Actual:    string(br.output),
			Error:     br.err,
		}
	}

	actual := strings.TrimRight(string(br.output), "\n")

	return &TestResult{
		Passed:    strings.Contains(actual, expected),
		StartLine: startLine,
		Expected:  expected,
		Actual:    actual,
	}
}

func outr(block *nodes.MakeDoCodeBlock, source []byte, br blockResult, startLine int) *TestResult {
	directive := block.GetDirective(nodes.DirectiveOutRegex)
	if directive == nil {
		return nil
	}

	expected := directive.ContentString(source)

	if br.err != nil {
		return &TestResult{
			Passed:    false,
			StartLine: startLine,
			Expected:  expected,
			Actual:    string(br.output),
			Error:     br.err,
		}
	}

	actual := strings.TrimRight(string(br.output), "\n")

	re, err := regexp.Compile(expected)
	if err != nil {
		return &TestResult{
			Passed:    false,
			StartLine: startLine,
			Expected:  expected,
			Actual:    actual,
			Error:     fmt.Errorf("invalid regex '%s': %w", expected, err),
		}
	}

	return &TestResult{
		Passed:    re.MatchString(actual),
		StartLine: startLine,
		Expected:  expected,
		Actual:    actual,
	}
}

// TODO: think how to use go routines to support background command like running a server.
func cmd(block *nodes.MakeDoCodeBlock, source []byte, br blockResult, startLine int) *TestResult {
	directive := block.GetDirective(nodes.DirectiveCmd)
	if directive == nil {
		return nil
	}

	// Code block must succeed first
	if br.err != nil {
		return &TestResult{
			Passed:    false,
			StartLine: startLine,
			Expected:  "code block to succeed",
			Actual:    string(br.output),
			Error:     br.err,
		}
	}

	// Run verification command
	command := directive.ContentString(source)
	verifyCmd := exec.Command(os.Getenv("SHELL"), "-c", command)
	err := verifyCmd.Run()

	return &TestResult{
		Passed:    err == nil,
		StartLine: startLine,
		Expected:  "exit 0",
		Actual:    command,
		Error:     err,
	}
}

func normalizePath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	return p
}

func matchGlob(path, pattern string) (bool, error) {
	starCount := strings.Count(pattern, "*")

	if starCount > 1 {
		return false, fmt.Errorf("pwd directive only allows a single wildcard (*) at the start or end of the pattern, got: %q", pattern)
	}

	if starCount == 1 {
		if strings.HasPrefix(pattern, "*") {
			return strings.HasSuffix(path, strings.TrimPrefix(pattern, "*")), nil
		}
		if strings.HasSuffix(pattern, "*") {
			return strings.HasPrefix(path, strings.TrimSuffix(pattern, "*")), nil
		}

		// * is in the middle
		return false, fmt.Errorf("pwd directive only allows a wildcard (*) at the start or end of the pattern, got: %q", pattern)
	}

	// Default: match suffix
	return strings.HasSuffix(path, pattern), nil
}

func pwd(block *nodes.MakeDoCodeBlock, source []byte, br blockResult, startLine int) *TestResult {
	directive := block.GetDirective(nodes.DirectivePwd)
	if directive == nil {
		return nil
	}

	pattern := strings.TrimSpace(directive.ContentString(source))

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return &TestResult{
			Passed:    false,
			StartLine: startLine,
			Expected:  pattern,
			Actual:    "",
			Error:     fmt.Errorf("failed to get working directory: %w", err),
		}
	}

	// Normalize both pattern and cwd
	cwd = normalizePath(cwd)
	pattern = normalizePath(pattern)

	// Match using glob rules
	matched, err := matchGlob(cwd, pattern)
	if err != nil {
		return &TestResult{
			Passed:    false,
			StartLine: startLine,
			Expected:  pattern,
			Actual:    "",
			Error:     err,
		}
	}

	return &TestResult{
		Passed:    matched,
		StartLine: startLine,
		Expected:  pattern,
		Actual:    cwd,
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

		// Execute block once
		lines := block.Lines()
		startLine := lineNumber(source, lines.At(0).Start)
		br := execBlock(block, source)

		// Process all directives on this block
		directives := block.Directives()
		if len(directives) == 0 {
			return ast.WalkContinue, nil
		}

		for _, directive := range directives {
			var result *TestResult

			switch directive.Kind {
			case nodes.DirectiveOut:
				result = out(block, source, br, startLine)
			case nodes.DirectiveOutRegex:
				result = outr(block, source, br, startLine)
			case nodes.DirectiveCmd:
				result = cmd(block, source, br, startLine)
			case nodes.DirectivePwd:
				result = pwd(block, source, br, startLine)
			}

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
