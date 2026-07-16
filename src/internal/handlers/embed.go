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

func EmbedMarkdownFile(mdFile string, ctx *engine.RunContext) error {
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

		code := string(block.Code(source))
		lineNum := lineNumber(source, block.Lines().At(0).Start)

		outcome := engine.EvaluateBlock(code, directives, source, lineNum, ctx)

		for _, warn := range outcome.ExecResult.Warnings {
			allWarnings = append(allWarnings, fmt.Sprintf("%s:%d: %v", mdFile, lineNum, warn))
		}

		var prefix, oldOutputBytes, newOutputBytes, suffix []byte
		var nextPos int

		if block.IsConsoleFormat() {
			blockStart := lines.At(0).Start
			prefix = make([]byte, 0, (blockStart-lastPos)+lines.Len()*20)
			prefix = append(prefix, source[lastPos:blockStart]...)

			var oldOutput bytes.Buffer
			inCommands := true
			for i := 0; i < lines.Len(); i++ {
				seg := lines.At(i)
				line := seg.Value(source)
				trimmed := bytes.TrimSpace(line)
				if inCommands && nodes.IsConsoleCommand(trimmed) {
					prefix = append(prefix, line...)
				} else {
					inCommands = false
					oldOutput.Write(line)
				}
			}

			oldOutputBytes = oldOutput.Bytes()

			finalOutput := outcome.FinalOutput
			if len(finalOutput) > 0 {
				newOutputBytes = append([]byte(nil), finalOutput...)
				if finalOutput[len(finalOutput)-1] != '\n' {
					newOutputBytes = append(newOutputBytes, '\n')
				}
			}

			suffix = source[lastLineEnd:blockEnd]
			nextPos = blockEnd

			if outBlock := block.OutputBlock(); outBlock != nil {
				end := outBlock.Lines().At(outBlock.Lines().Len() - 1).Stop
				nextPos = end + bytes.IndexByte(source[end:], '\n') + 1
			}
		} else {
			prefix = source[lastPos:blockEnd]
			nextPos = blockEnd

			if outBlock := block.OutputBlock(); outBlock != nil {
				start := outBlock.Lines().At(0).Start
				start = bytes.LastIndexByte(source[:start-1], '\n') + 1
				end := outBlock.Lines().At(outBlock.Lines().Len() - 1).Stop
				end = end + bytes.IndexByte(source[end:], '\n') + 1
				oldOutputBytes = source[start:end]
				nextPos = end
			}

			finalOutput := string(outcome.FinalOutput)
			if len(finalOutput) > 0 || block.OutputBlock() != nil {
				newOutputBytes = []byte(fmt.Sprintf("\n```stdout\n%s\n```\n", finalOutput))
			}
		}

		if !outcome.Passed {
			failReason := outcome.FailReason.Error()
			fmt.Printf("[WARN] Block at line %d failed (%s), skipping output update\n", lineNum, failReason)
			failed++

			buf = append(buf, prefix...)
			if len(oldOutputBytes) > 0 {
				buf = append(buf, oldOutputBytes...)
			}
			buf = append(buf, suffix...)
		} else {
			buf = append(buf, prefix...)

			isUnchanged := bytes.Equal(bytes.TrimSpace(oldOutputBytes), bytes.TrimSpace(newOutputBytes))

			if len(oldOutputBytes) > 0 {
				if isUnchanged {
					unchanged++
					buf = append(buf, oldOutputBytes...)
				} else {
					updated++
					buf = append(buf, newOutputBytes...)
				}
			} else {
				if len(newOutputBytes) > 0 {
					added++
					buf = append(buf, newOutputBytes...)
				}
			}

			buf = append(buf, suffix...)
		}

		lastPos = nextPos
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
