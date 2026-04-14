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
	codeBlock   *ast.FencedCodeBlock
	htmlBlocks  []*ast.HTMLBlock
	directives  []*Directive
	outputBlock *ast.FencedCodeBlock
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

		// Don't turn stdout blocks into MakeDoCodeBlocks.
		// They are the output of previous MakeDoCodeBlocks.
		lang := codeBlock.Language(source)
		if bytes.Equal(lang, []byte("stdout")) {
			return ast.WalkContinue, nil
		}

		var directives []*Directive
		var htmlBlocks []*ast.HTMLBlock
		var next ast.Node

		// Collect consecutive HTML blocks that are valid directives
		for next = codeBlock.NextSibling(); next != nil; next = next.NextSibling() {
			htmlBlock, ok := next.(*ast.HTMLBlock)
			if !ok {
				break
			}

			content := extractHTMLBlockContent(htmlBlock, source)
			if content == nil {
				break
			}

			var offset int
			if htmlBlock.Lines().Len() > 0 {
				offset = htmlBlock.Lines().At(0).Start
			}

			directive, ok := ParseDirective(content, offset)
			if !ok {
				break
			}

			directives = append(directives, directive)
			htmlBlocks = append(htmlBlocks, htmlBlock)
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
			codeBlock:   codeBlock,
			htmlBlocks:  htmlBlocks,
			directives:  directives,
			outputBlock: outputBlock,
		})

		return ast.WalkContinue, nil
	})

	// Apply transformations
	for _, tr := range transformations {
		makedo := NewMakeDoCodeBlock(tr.codeBlock)
		for _, d := range tr.directives {
			makedo.AddDirective(d)
		}
		if tr.outputBlock != nil {
			makedo.SetOutputBlock(tr.outputBlock)
		}

		parent := tr.codeBlock.Parent()
		parent.ReplaceChild(parent, tr.codeBlock, makedo)
		for _, htmlBlock := range tr.htmlBlocks {
			parent.RemoveChild(parent, htmlBlock)
		}
		if tr.outputBlock != nil {
			parent.RemoveChild(parent, tr.outputBlock)
		}
	}
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
