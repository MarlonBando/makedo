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

func EmbedMarkdownFile(mdFile string, ctx *engine.RunContext) error {
	mdFile = strings.TrimSpace(mdFile)

	source, err := os.ReadFile(mdFile)
	if err != nil {
		return err
	}

	var allWarnings []string

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

		code := string(block.Code(source))
		lineNum := lineNumber(source, block.Lines().At(0).Start)

		outcome := engine.EvaluateBlock(code, directives, source, lineNum, ctx)

		for _, warn := range outcome.ExecResult.Warnings {
			allWarnings = append(allWarnings, fmt.Sprintf("%s:%d: %v", mdFile, lineNum, warn))
		}

		if !outcome.Passed {
			failReason := outcome.FailReason.Error()
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

		// Use the FinalOutput calculated by the EvaluateBlock which already handled Substitutions
		finalOutput := string(outcome.FinalOutput)

		var newOutput string
		if len(finalOutput) > 0 || len(directives) > 0 {
			newOutput = fmt.Sprintf("\n```stdout\n%s\n```\n", finalOutput)
		} else {
			// If there's literally no output and no directives, we probably don't even need a stdout block?
			// But for consistency let's preserve the original behavior which printed it if string(finalOutput) wasn't completely empty after trim.
			// Let's just follow the original logic which allowed empty blocks if we appended them.
			newOutput = fmt.Sprintf("\n```stdout\n%s\n```\n", finalOutput)
		}

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
		} else if len(finalOutput) > 0 {
			added++
			buf = append(buf, newOutput...)
			lastPos = blockEnd
		} else {
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
