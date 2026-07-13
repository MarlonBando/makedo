package nodes

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// transformation holds data for a pending AST transformation.
type transformation struct {
	codeBlock     *ast.FencedCodeBlock
	nodesToRemove []ast.Node
	directives    []*Directive
	outputBlock   *ast.FencedCodeBlock
}

type MakeDoTransformer struct{}

// Scan the AST for FenceCodeBlock followed by an HTML comment block.
// If the comment is a directive (keyword + content), replace both nodes with a MakeDoCodeBlock.
func (t *MakeDoTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()

	// we collect the transformation to apply after the walking.
	// can't transform while walking the ast
	transformations := make([]transformation, 0, 8)

	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		codeBlock, ok := n.(*ast.FencedCodeBlock)
		if !ok {
			return ast.WalkContinue, nil
		}

		if codeBlock.Lines().Len() == 0 {
			return ast.WalkContinue, nil
		}

		lang := codeBlock.Language(source)
		if !isShellLanguage(lang) {
			return ast.WalkContinue, nil
		}

		var directives []*Directive
		var nodesToRemove []ast.Node
		hasSkip := false
		var next ast.Node

		// Collect consecutive HTML blocks (or html blocks in codespan) that are valid directives
		for next = codeBlock.NextSibling(); next != nil; next = next.NextSibling() {
			content, start, isDirective := getDirectiveContent(next, source)
			if !isDirective {
				break
			}

			directive, ok := ParseDirective(content, start)
			if !ok {
				break
			}

			if directive.Kind == DirectiveSkip {
				hasSkip = true
			}

			directives = append(directives, directive)
			nodesToRemove = append(nodesToRemove, next)
		}

		if hasSkip {
			return ast.WalkContinue, nil
		}

		var outputBlock *ast.FencedCodeBlock
		if next != nil {
			if fb, ok := next.(*ast.FencedCodeBlock); ok {
				lang := fb.Language(source)
				if bytes.Equal(lang, []byte("stdout")) {
					outputBlock = fb
				}
			}
		}

		transformations = append(transformations, transformation{
			codeBlock:     codeBlock,
			nodesToRemove: nodesToRemove,
			directives:    directives,
			outputBlock:   outputBlock,
		})

		return ast.WalkContinue, nil
	})

	// Apply transformations
	for _, tr := range transformations {
		makedo := NewMakeDoCodeBlock(tr.codeBlock, source)
		for _, d := range tr.directives {
			makedo.AddDirective(d)
		}
		if tr.outputBlock != nil {
			makedo.SetOutputBlock(tr.outputBlock)
		}

		parent := tr.codeBlock.Parent()
		parent.ReplaceChild(parent, tr.codeBlock, makedo)
		for _, htmlBlock := range tr.nodesToRemove {
			parent.RemoveChild(parent, htmlBlock)
		}
		if tr.outputBlock != nil {
			parent.RemoveChild(parent, tr.outputBlock)
		}
	}
}

func getDirectiveContent(n ast.Node, source []byte) ([]byte, int, bool) {
	var start int
	if htmlBlock, ok := n.(*ast.HTMLBlock); ok {
		if htmlBlock.Lines().Len() > 0 {
			start = htmlBlock.Lines().At(0).Start
		}
		return extractHTMLBlockContent(htmlBlock, source), start, true
	}

	paragraph, ok := n.(*ast.Paragraph)
	if !ok {
		return nil, start, false
	}
	if paragraph.ChildCount() != 1 {
		return nil, start, false
	}
	codeSpan, ok := paragraph.FirstChild().(*ast.CodeSpan)
	if !ok {
		return nil, start, false
	}

	codeSpanChild := codeSpan.FirstChild()
	if codeSpanChild == nil {
		return nil, start, false
	}

	textNode, ok := codeSpanChild.(*ast.Text)
	if !ok {
		return nil, 0, false
	}

	start = textNode.Segment.Start
	content := bytes.TrimSpace(textNode.Segment.Value(source))

	if !bytes.HasPrefix(content, []byte("<!--")) || !bytes.HasSuffix(content, []byte("-->")) {

		return nil, 0, false
	}

	return content, start, true
}

func extractHTMLBlockContent(block *ast.HTMLBlock, source []byte) []byte {
	lines := block.Lines()
	if lines.Len() == 0 {
		return nil
	}

	if lines.Len() == 1 {
		seg := lines.At(0)
		return seg.Value(source)
	}

	// Multi-line: collect all lines
	var buf bytes.Buffer
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(seg.Value(source))
	}
	return buf.Bytes()
}

func isShellLanguage(lang []byte) bool {
	return bytes.Equal(lang, []byte("bash")) ||
		bytes.Equal(lang, []byte("sh")) ||
		bytes.Equal(lang, []byte("zsh")) ||
		bytes.Equal(lang, []byte("shell"))
}

// MakeDoExtension is a goldmark extension that transforms code blocks with
// directive comments into MakeDoCodeBlock nodes.
type MakeDoExtension struct{}

// NewMakeDoExtension creates a new MakeDoExtension.
// Directive keywords are defined statically in directive_kind.go.
func NewMakeDoExtension() *MakeDoExtension {
	return &MakeDoExtension{}
}

func (e *MakeDoExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&MakeDoTransformer{}, 100),
		),
	)
}
