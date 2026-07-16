package handlers

import (
	"bytes"
	"fmt"
	"os"

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

	var passedTests int
	var failedTests int
	var failedBlocks = make([]*failedBlock, 0, bytes.Count(source, []byte("\n```"))/2)

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

		progressBlockHeader(mdPath, startLine)
		defer progressBlockEnd()

		for _, warn := range outcome.ExecResult.Warnings {
			allWarnings = append(allWarnings, fmt.Sprintf("%s:%d: %v", mdPath, startLine, warn))
		}

		// Directiveless block: counts as a single test that checks the exit code.
		if len(directives) == 0 {
			if outcome.Passed {
				passedTests++
				progressMark(true)
			} else {
				failedTests++
				progressMark(false)
				failedBlocks = append(failedBlocks, &failedBlock{
					mdPath:    mdPath,
					startLine: startLine,
					block:     block,
					outcome:   outcome,
				})
			}
			return ast.WalkContinue, nil
		}

		if commandFailed(outcome) {
			// Count the whole block as a single failure.
			failedTests++
			progressMark(false)
			failedBlocks = append(failedBlocks, &failedBlock{
				mdPath:    mdPath,
				startLine: startLine,
				block:     block,
				outcome:   outcome,
			})
			return ast.WalkContinue, nil
		}

		blockHasFailure := false
		for _, tr := range outcome.TestResults {
			if tr.Passed {
				passedTests++
				progressMark(true)
			} else {
				failedTests++
				progressMark(false)
				blockHasFailure = true
			}
		}
		if blockHasFailure {
			failedBlocks = append(failedBlocks, &failedBlock{
				mdPath:    mdPath,
				startLine: startLine,
				block:     block,
				outcome:   outcome,
			})
		}

		return ast.WalkContinue, nil
	})

	if walkErr != nil {
		return walkErr
	}

	// Failure panels (above summary)
	if len(failedBlocks) > 0 {
		fmt.Println()
		fmt.Println()
		for _, fb := range failedBlocks {
			fmt.Println(renderFailurePanel(fb, source))
		}
	}

	// Warnings
	if len(allWarnings) > 0 {
		fmt.Println("Warnings:")
		for _, w := range allWarnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	// Summary
	total := passedTests + failedTests
	renderSummary(passedTests, total, len(failedBlocks))

	if failedTests > 0 {
		return fmt.Errorf("%d test(s) failed", failedTests)
	}
	return nil
}
