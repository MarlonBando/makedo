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

func EmbedMarkdownFile(mdFile string) error {
	mdFile = strings.TrimSpace(mdFile)

	source, err := os.ReadFile(mdFile)
	if err != nil {
		return err
	}

	var allWarnings []string

	// Process registry for cleanup at document end
	registry := executor.NewRegistry()
	defer registry.KillAll()

	md := goldmark.New(
		goldmark.WithExtensions(
			nodes.NewMakeDoExtension(),
		),
	)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var blocks []*nodes.MakeDoCodeBlock

	// Collect all MakeDoCodeBlocks
	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if block, ok := n.(*nodes.MakeDoCodeBlock); ok {
			blocks = append(blocks, block)
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return err
	}

	buf := make([]byte, 0, len(source)+len(blocks)*128)
	lastPos := 0

	var added, updated, unchanged, failed int

	for _, block := range blocks {
		// Calculate the end of the block (including directives)
		var blockEnd int
		directives := block.Directives()
		lines := block.Lines()
		lastLineEnd := lines.At(lines.Len() - 1).Stop

		// Let's find the actual end of the directives by searching for the last "-->" after the code block
		blockEnd = lastLineEnd + bytes.IndexByte(source[lastLineEnd:], '\n') + 1

		if len(directives) > 0 {
			// Find the last "-->" after the code block
			// We can bound the search to either EOF or the start of the next node/output block
			searchArea := source[blockEnd:]
			var lastDirectiveEnd int
			for range directives {
				idx := bytes.Index(searchArea[lastDirectiveEnd:], []byte("-->"))
				if idx != -1 {
					lastDirectiveEnd += idx + 3 // length of "-->"
				}
			}
			if lastDirectiveEnd > 0 {
				blockEnd += lastDirectiveEnd
				// include the trailing newline if present
				if blockEnd < len(source) && source[blockEnd] == '\n' {
					blockEnd++
				}
			}
		}

		buf = append(buf, source[lastPos:blockEnd]...)

		patterns, err := executor.PrecompileDirectives(directives, source)
		if err != nil {
			fmt.Printf("[WARN] Failed to precompile directives: %v\n", err)
			continue
		}

		code := string(block.Code(source))
		result := executor.Execute(code, directives, source, false)
		lineNum := lineNumber(source, block.Lines().At(0).Start)

		for _, warn := range result.Warnings {
			allWarnings = append(allWarnings, fmt.Sprintf("%s:%d: %v", mdFile, lineNum, warn))
		}

		if result.Process != nil && result.Status != executor.Completed {
			registry.Add(result.Process)
		}

		blockFailed := false
		var failReason string

		if result.Err != nil {
			blockFailed = true
			failReason = result.Err.Error()
		} else if result.Status == executor.Completed && result.ExitCode > 0 {
			blockFailed = true
			failReason = fmt.Sprintf("exit code %d", result.ExitCode)
		} else if result.Status == executor.Stalled && len(directives) > 0 {
			blockFailed = true
			failReason = "command stalled before directives passed"
		} else {
			for _, d := range directives {
				testRes := testDirective(d, result, source, lineNum, patterns)
				if testRes != nil && !testRes.Passed {
					blockFailed = true
					if testRes.Error != nil {
						failReason = testRes.Error.Error()
					} else {
						failReason = fmt.Sprintf("directive '%s' failed to match", string(d.Kind))
					}
					break
				}
			}
		}

		if blockFailed {
			fmt.Printf("[WARN] Block at line %d failed (%s), skipping output update\n", lineNum, failReason)
			failed++

			// Keep existing output block if present
			if outBlock := block.OutputBlock(); outBlock != nil {
				start := outBlock.Lines().At(0).Start
				start = bytes.LastIndexByte(source[:start-1], '\n') + 1
				end := outBlock.Lines().At(outBlock.Lines().Len() - 1).Stop
				end = end + bytes.IndexByte(source[end:], '\n') + 1
				buf = append(buf, source[start:end]...)
				lastPos = end
			} else {
				lastPos = blockEnd
			}
			continue
		}

		// Substitute patterns in output string
		finalOutput := bytes.TrimSpace(result.Output)

		if !blockFailed && len(directives) > 0 {
			finalOutput = executor.SubstituteOutput(finalOutput, directives, source)
		}

		newOutput := fmt.Sprintf("\n```stdout\n%s\n```\n", string(finalOutput))

		if outBlock := block.OutputBlock(); outBlock != nil {
			// Extract old output content
			start := outBlock.Lines().At(0).Start
			start = bytes.LastIndexByte(source[:start-1], '\n') + 1
			end := outBlock.Lines().At(outBlock.Lines().Len() - 1).Stop
			end = end + bytes.IndexByte(source[end:], '\n') + 1

			oldOutput := string(source[start:end])

			if strings.TrimSpace(oldOutput) == strings.TrimSpace(newOutput) {
				unchanged++
				buf = append(buf, oldOutput...) // preserve exact original if unchanged
			} else {
				updated++
				buf = append(buf, newOutput...)
			}
			lastPos = end
		} else if len(newOutput) > 0 {
			added++
			buf = append(buf, newOutput...)
			lastPos = blockEnd
		}
	}

	// Write any remaining text
	buf = append(buf, source[lastPos:]...)

	// Atomic write
	tmpFile := mdFile + ".tmp"
	if err := os.WriteFile(tmpFile, buf, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, mdFile); err != nil {
		os.Remove(tmpFile)
		return err
	}

	summary := fmt.Sprintf("Embed complete: %d added, %d updated, %d unchanged", added, updated, unchanged)
	if failed > 0 {
		summary += fmt.Sprintf(", %d failed", failed)
	}
	fmt.Println(summary)

	if len(allWarnings) > 0 {
		fmt.Println("\nWarnings Summary:")
		for _, w := range allWarnings {
			fmt.Printf("- %s\n", w)
		}
	}

	return nil
}
