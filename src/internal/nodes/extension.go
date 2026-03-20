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
	codeBlock *ast.FencedCodeBlock
	htmlBlock *ast.HTMLBlock
	directive *Directive
}

type MakeDoTransformer struct {
	Registry *Registry
}

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

		next := codeBlock.NextSibling()
		if next == nil {
			return ast.WalkContinue, nil
		}

		// Markdown comments are parsed into HTML blocked by goldmark
		htmlBlock, ok := next.(*ast.HTMLBlock)
		if !ok {
			return ast.WalkContinue, nil
		}

		content := extractHTMLBlockContent(htmlBlock, source)
		if content == nil {
			return ast.WalkContinue, nil
		}

		var offset int
		if htmlBlock.Lines().Len() > 0 {
			offset = htmlBlock.Lines().At(0).Start
		}

		directive, ok := ParseDirective(content, offset, t.Registry)
		if !ok {
			return ast.WalkContinue, nil
		}

		transformations = append(transformations, transformation{
			codeBlock: codeBlock,
			htmlBlock: htmlBlock,
			directive: directive,
		})

		return ast.WalkContinue, nil
	})

	// Apply transformations
	for _, tr := range transformations {
		makedo := NewMakeDoCodeBlock(tr.codeBlock)
		makedo.AddDirective(tr.directive)

		parent := tr.codeBlock.Parent()
		parent.ReplaceChild(parent, tr.codeBlock, makedo)
		parent.RemoveChild(parent, tr.htmlBlock)
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

//NOTE: The extension part with inside the registration of the keyword is a bit "meh"
// extension is 100% neeeded because it tells goldmark to use our custome node
// but the register keyword there it looks a bit weird at first
// TODO: evaluate if it's a good solution or there are better approach

type MakeDoExtension struct {
	Registry *Registry
}

func NewMakeDoExtension() *MakeDoExtension {
	r := NewRegistry()
	r.Register("out")
	return &MakeDoExtension{Registry: r}
}

func NewMakeDoExtensionWithKeywords(keywords ...string) *MakeDoExtension {
	r := NewRegistry()
	for _, k := range keywords {
		r.Register(k)
	}
	return &MakeDoExtension{Registry: r}
}

func (e *MakeDoExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&MakeDoTransformer{Registry: e.Registry}, 100),
		),
	)
}
