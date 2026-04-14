package handlers

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"makedo/internal/executor"
	"makedo/internal/nodes"

	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var warnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))    // Yellow
var successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green

func EmbedMarkdownFile(mdFile string) error {
	mdFile = strings.TrimSpace(mdFile)

	source, err := os.ReadFile(mdFile)
	if err != nil {
		return err
	}

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

	var buf bytes.Buffer
	buf.Grow(len(source) + len(blocks)*128)
	lastPos := 0

	var added, updated, unchanged, failed int

	for _, block := range blocks {
		// Calculate the end of the block (including directives)
		var blockEnd int
		directives := block.Directives()
		lines := block.Lines()
		lastLineEnd := lines.At(lines.Len() - 1).Stop

		// The issue with lastDir.End is that the Directive struct does not store the end position of the directive block itself,
		// only the content segment. We can find the end of the last directive by scanning the source after the code block.
		// Alternatively, since MakeDoCodeBlock replaced the code block AND the directives in the AST,
		// and we know the last HTML block was removed, its byte range in the original source is everything up to the next sibling,
		// OR we can just use the start of the outputBlock if it exists.

		// Let's find the actual end of the directives by searching for the last "-->" after the code block
		blockEnd = lastLineEnd + bytes.IndexByte(source[lastLineEnd:], '\n') + 1

		if len(directives) > 0 {
			// Find the last "-->" after the code block
			// We can bound the search to either EOF or the start of the next node/output block
			searchArea := source[blockEnd:]
			var lastDirectiveEnd int
			for i := 0; i < len(directives); i++ {
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

		// Write everything from lastPos to the end of the block
		buf.Write(source[lastPos:blockEnd])

		code := string(block.Code(source))
		result := executor.Execute(code, directives, source, false)

		if result.Process != nil && result.Status != executor.Completed {
			registry.Add(result.Process)
		}

		lineNum := lineNumber(source, block.Lines().At(0).Start)

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
				testRes := testDirective(d, result, source, lineNum)
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
			fmt.Println(warnStyle.Render(fmt.Sprintf("[WARN] Block at line %d failed (%s), skipping output update", lineNum, failReason)))
			failed++

			// Keep existing output block if present
			if outBlock := block.OutputBlock(); outBlock != nil {
				start := outBlock.Lines().At(0).Start
				start = bytes.LastIndexByte(source[:start-1], '\n') + 1
				end := outBlock.Lines().At(outBlock.Lines().Len() - 1).Stop
				end = end + bytes.IndexByte(source[end:], '\n') + 1
				buf.Write(source[start:end])
				lastPos = end
			} else {
				lastPos = blockEnd
			}
			continue
		}

		newOutput := fmt.Sprintf("\n```stdout\n%s\n```\n", string(bytes.TrimSpace(result.Output)))

		if outBlock := block.OutputBlock(); outBlock != nil {
			// Extract old output content
			start := outBlock.Lines().At(0).Start
			start = bytes.LastIndexByte(source[:start-1], '\n') + 1
			end := outBlock.Lines().At(outBlock.Lines().Len() - 1).Stop
			end = end + bytes.IndexByte(source[end:], '\n') + 1

			oldOutput := string(source[start:end])

			if strings.TrimSpace(oldOutput) == strings.TrimSpace(newOutput) {
				unchanged++
				buf.WriteString(oldOutput) // preserve exact original if unchanged
			} else {
				updated++
				buf.WriteString(newOutput)
			}
			lastPos = end
		} else {
			added++
			buf.WriteString(newOutput)
			lastPos = blockEnd
		}
	}

	// Write any remaining text
	buf.Write(source[lastPos:])

	// Atomic write
	tmpFile := mdFile + ".tmp"
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0644); err != nil {
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
	fmt.Println(successStyle.Render(summary))

	return nil
}
