package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"makedo/internal/nodes"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red

func getNoIndentStyle() ansi.StyleConfig {
	style := styles.DarkStyleConfig
	style.Document.Margin = uintPtr(0)
	return style
}

func uintPtr(i uint) *uint {
	return &i
}

// renderContext holds state for sequential rendering
type renderContext struct {
	source          []byte
	glamourRenderer *glamour.TermRenderer
	lastPos         int
}

func RunMarkdownFile(mdFile string) error {
	mdFile = strings.TrimSpace(mdFile)

	source, err := os.ReadFile(mdFile)
	if err != nil {
		return err
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(getNoIndentStyle()),
		glamour.WithWordWrap(0),
	)
	if err != nil {
		return err
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			nodes.NewMakeDoExtension(),
		),
	)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	ctx := &renderContext{
		source:          source,
		glamourRenderer: renderer,
		lastPos:         0,
	}

	// Walk AST and render/execute sequentially
	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case nodes.KindMakeDoCodeBlock:
			return ctx.handleMakeDoCodeBlock(n.(*nodes.MakeDoCodeBlock))
		case ast.KindFencedCodeBlock:
			return ctx.handleFencedCodeBlock(n.(*ast.FencedCodeBlock))
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return err
	}

	// Render any remaining markdown after last code block
	if ctx.lastPos < len(source) {
		remaining := source[ctx.lastPos:]
		if len(strings.TrimSpace(string(remaining))) > 0 {
			rendered, err := ctx.glamourRenderer.Render(string(remaining))
			if err != nil {
				return err
			}
			fmt.Print(rendered)
		}
	}

	return nil
}

// getCodeBlockFullRange returns the byte range including fence markers
func getCodeBlockFullRange(node ast.Node, source []byte) (start, end int) {
	lines := node.Lines()
	if lines.Len() == 0 {
		return 0, 0
	}

	// Get content boundaries
	firstContentLine := lines.At(0)
	lastContentLine := lines.At(lines.Len() - 1)
	contentStart := firstContentLine.Start
	contentEnd := lastContentLine.Stop

	// Find opening fence: scan backwards from contentStart to find the fence line
	// contentStart points to the first line of code content
	// The fence line (```bash or similar) is on the line before that
	fenceStart := contentStart
	// First, scan back past any newline immediately before content
	if fenceStart > 0 && source[fenceStart-1] == '\n' {
		fenceStart--
	}
	// Now scan back to the previous newline (start of fence line)
	for fenceStart > 0 && source[fenceStart-1] != '\n' {
		fenceStart--
	}

	// Find closing fence by scanning forward from content end
	fenceEnd := contentEnd
	// Skip to end of closing fence line (find next newline)
	for fenceEnd < len(source) && source[fenceEnd] != '\n' {
		fenceEnd++
	}
	// Include the newline after closing fence
	if fenceEnd < len(source) {
		fenceEnd++
	}

	return fenceStart, fenceEnd
}

func (ctx *renderContext) handleMakeDoCodeBlock(block *nodes.MakeDoCodeBlock) (ast.WalkStatus, error) {
	return ctx.handleCodeBlock(block, block.Code(ctx.source))
}

func (ctx *renderContext) handleFencedCodeBlock(codeNode *ast.FencedCodeBlock) (ast.WalkStatus, error) {
	var code strings.Builder
	for i := 0; i < codeNode.Lines().Len(); i++ {
		line := codeNode.Lines().At(i)
		code.Write(line.Value(ctx.source))
	}
	return ctx.handleCodeBlock(codeNode, []byte(code.String()))
}

func (ctx *renderContext) handleCodeBlock(node ast.Node, code []byte) (ast.WalkStatus, error) {
	start, end := getCodeBlockFullRange(node, ctx.source)

	// Render markdown content before this code block
	if ctx.lastPos < start {
		section := ctx.source[ctx.lastPos:start]
		if len(strings.TrimSpace(string(section))) > 0 {
			rendered, err := ctx.glamourRenderer.Render(string(section))
			if err != nil {
				return ast.WalkStop, err
			}
			fmt.Print(rendered)
		}
	}

	// Render the code block itself (syntax highlighted)
	codeBlockMarkdown := ctx.source[start:end]
	rendered, err := ctx.glamourRenderer.Render(string(codeBlockMarkdown))
	if err != nil {
		return ast.WalkStop, err
	}
	fmt.Print(rendered)

	// Execute with streaming output
	fmt.Println("---")

	cmd := exec.Command(os.Getenv("SHELL"), "-c", string(code))
	cmd.Stdout = os.Stdout // Stream directly
	cmd.Stderr = os.Stderr // Stream directly

	execErr := cmd.Run()

	fmt.Println("---")

	// Print error message if execution failed
	if execErr != nil {
		exitCode := "unknown"
		if exitErr, ok := execErr.(*exec.ExitError); ok {
			exitCode = fmt.Sprintf("%d", exitErr.ExitCode())
		}
		fmt.Println(errorStyle.Render(fmt.Sprintf("[ERROR] exit code %s", exitCode)))
	}

	// Update position tracker
	ctx.lastPos = end

	return ast.WalkContinue, nil
}
