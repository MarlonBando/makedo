package handlers

import (
	"bytes"
	"fmt"
	"strings"

	"makedo/internal/engine"
	"makedo/internal/nodes"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

// Colors. lipgloss auto-downgrades on non-TTY / NO_COLOR.
var (
	passStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
	failStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // red
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // grey
	headerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	boxStyle     = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("9")).
			PaddingLeft(1)
	summaryOK   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	summaryFail = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
)

// failedBlock is what we collect during the walk to render after.
type failedBlock struct {
	mdPath    string
	startLine int
	block     *nodes.MakeDoCodeBlock
	outcome   *engine.BlockOutcome
	isSetup   bool // no directives
}

// codeRenderer is a glamour renderer stripped of outer padding/margins,
// used to syntax-highlight the fenced code inside a panel.
// Falls back to nil if construction fails; renderCode handles nil.
var codeRenderer *glamour.TermRenderer

func init() {
	codeRenderer = buildCodeRenderer()
}

func buildCodeRenderer() *glamour.TermRenderer {
	style := styles.DarkStyleConfig
	style.Document.Margin = uintPtr(0)
	style.Document.BlockPrefix = ""
	style.Document.BlockSuffix = ""
	if style.CodeBlock.Margin == nil {
		style.CodeBlock.Margin = uintPtr(0)
	} else {
		*style.CodeBlock.Margin = 0
	}
	// (Chroma styling for syntax tokens left as-is; we only need to flatten outer block spacing.)
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(0),
	)
	if err != nil {
		return nil
	}
	return r
}

// renderCode returns syntax-highlighted code, or plain code on failure.
func renderCode(code, lang string) string {
	code = strings.TrimRight(code, "\n")
	if codeRenderer == nil {
		return code
	}
	md := "```" + lang + "\n" + code + "\n```"
	out, err := codeRenderer.Render(md)
	if err != nil {
		return code
	}
	return strings.TrimRight(out, "\n")
}

// classifyBlock returns true if the command itself failed (so directives
// were never meaningfully evaluated against patterns).
func commandFailed(o *engine.BlockOutcome) bool {
	if o == nil || o.ExecResult == nil {
		return false
	}
	r := o.ExecResult
	if r.Err != nil {
		return true
	}
	if r.Status == engine.Stalled {
		return true
	}
	if r.Status == engine.Completed && r.ExitCode != 0 {
		return true
	}
	return false
}

// renderFailurePanel writes a single failure panel to a string.
func renderFailurePanel(fb *failedBlock, source []byte) string {
	var sb strings.Builder

	// Header
	header := fmt.Sprintf("FAIL %s:%d", fb.mdPath, fb.startLine)
	if fb.isSetup {
		header += "  (setup)"
	}
	sb.WriteString(headerStyle.Render(header))
	sb.WriteString("\n")

	// Code box
	lang := string(fb.block.Language(source))
	code := string(fb.block.Code(source))
	rendered := renderCode(code, lang)
	sb.WriteString(boxStyle.Render(rendered))
	sb.WriteString("\n")

	cmdFail := commandFailed(fb.outcome)

	// Directives
	directives := fb.block.Directives()
	if len(directives) > 0 {
		// Map directive -> TestResult (positional; TestResults order matches directives
		// minus any skip/special, but our render only needs per-directive state, so we
		// pair by index where possible).
		results := fb.outcome.TestResults
		for i, d := range directives {
			line := string(source[d.Range.Start:d.Range.Stop])
			var styled string
			switch {
			case cmdFail:
				styled = dimStyle.Render("  " + line)
			case i < len(results) && results[i].Passed:
				styled = passStyle.Render("  ✓ " + line)
			case i < len(results):
				styled = failStyle.Render("  ✗ " + line)
			default:
				styled = dimStyle.Render("  " + line)
			}
			sb.WriteString(styled)
			sb.WriteString("\n")

			if !cmdFail && i < len(results) && !results[i].Passed {
				r := results[i]
				if r.Expected != "" {
					sb.WriteString(dimStyle.Render("      expected: " + oneLine(r.Expected)))
					sb.WriteString("\n")
				}
				if r.Actual != "" {
					sb.WriteString(dimStyle.Render("      actual:   " + oneLine(r.Actual)))
					sb.WriteString("\n")
				}
				if r.Error != nil {
					sb.WriteString(dimStyle.Render("      error:    " + r.Error.Error()))
					sb.WriteString("\n")
				}
			}
		}
	}

	// Command-failure reason line
	if cmdFail {
		reason := ""
		switch {
		case fb.outcome.ExecResult.Err != nil:
			reason = "error: " + fb.outcome.ExecResult.Err.Error()
		case fb.outcome.ExecResult.Status == engine.Stalled:
			reason = "stalled: no output before timeout"
		case fb.outcome.ExecResult.Status == engine.Completed:
			reason = fmt.Sprintf("exit %d", fb.outcome.ExecResult.ExitCode)
		}
		if reason != "" {
			sb.WriteString(failStyle.Render("  " + reason))
			sb.WriteString("\n")
		}
		// For setup blocks (or command failures), show tail of captured output.
		out := bytes.TrimSpace(fb.outcome.ExecResult.Output)
		if len(out) > 0 {
			sb.WriteString(dimStyle.Render("  --- output (tail) ---"))
			sb.WriteString("\n")
			for _, l := range tailLines(out, 20) {
				sb.WriteString(dimStyle.Render("  " + l))
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i] + " …"
	}
	if len(s) > 200 {
		return s[:200] + " …"
	}
	return s
}

func tailLines(b []byte, n int) []string {
	lines := strings.Split(string(b), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines
}

// progressMark prints a single streaming mark for a test result.
func progressMark(passed bool) {
	if passed {
		fmt.Print(passStyle.Render("."))
	} else {
		fmt.Print(failStyle.Render("F"))
	}
}

// renderSummary prints the final summary line.
// failedTests/totalTests are at the TestResult granularity; failedBlocks
// tells us how many distinct blocks failed (renderer-side dedupe).
func renderSummary(passed, total, failedBlocks int) {
	fmt.Println()
	fmt.Println()
	if failedBlocks == 0 {
		fmt.Println(summaryOK.Render(fmt.Sprintf("✓ %d/%d tests passed", passed, total)))
		return
	}
	fmt.Println(summaryFail.Render(
		fmt.Sprintf("✗ %d/%d tests passed — %d block(s) failed", passed, total, failedBlocks),
	))
}
