package handlers

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"makedo/internal/engine"
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
	runCtx          *engine.RunContext
}

func RunMarkdownFile(mdFile string, runCtx *engine.RunContext) error {
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
		runCtx:          runCtx,
	}

	// Walk AST and render/execute sequentially
	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case nodes.KindMakeDoCodeBlock:
			return ctx.handleMakeDoCodeBlock(n.(*nodes.MakeDoCodeBlock))
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

func (ctx *renderContext) handleMakeDoCodeBlock(block *nodes.MakeDoCodeBlock) (ast.WalkStatus, error) {
	return ctx.handleCodeBlock(block, block.Code(ctx.source), block.Directives())
}

func (ctx *renderContext) handleCodeBlock(node ast.Node, code []byte, directives []*nodes.Directive) (ast.WalkStatus, error) {
	lines := node.Lines()
	contentStart := lines.At(0).Start
	start := bytes.LastIndexByte(ctx.source[:contentStart-1], '\n') + 1
	lastLineEnd := lines.At(lines.Len() - 1).Stop
	end := lastLineEnd + bytes.IndexByte(ctx.source[lastLineEnd:], '\n') + 1

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

	result := engine.Execute(ctx.runCtx, string(code), directives, ctx.source, true)

	fmt.Println("---")

	// Register process for cleanup if still running
	if result.Process != nil && result.Status != engine.Completed {
		ctx.runCtx.Registry.Add(result.Process)
	}

	// Print error/status message
	if result.Err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("[ERROR] %v", result.Err)))
	} else if result.ExitCode > 0 {
		fmt.Println(errorStyle.Render(fmt.Sprintf("[ERROR] exit code %d", result.ExitCode)))
	}

	ctx.lastPos = end

	return ast.WalkContinue, nil
}
