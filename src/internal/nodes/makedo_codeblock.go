package nodes

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// KindMakeDoCodeBlock is the NodeKind for MakeDoCodeBlock.
var KindMakeDoCodeBlock = ast.NewNodeKind("MakeDoCodeBlock")

// MakeDoCodeBlock represents a fenced code block with associated directives.
// It combines a code block with parsed comment directives that follow it.
type MakeDoCodeBlock struct {
	ast.BaseBlock
	language    []byte
	info        *ast.Text
	directives  []*Directive
	outputBlock *ast.FencedCodeBlock
}

// NewMakeDoCodeBlock creates a MakeDoCodeBlock from a FencedCodeBlock.
// It copies the essential properties and line segments.
func NewMakeDoCodeBlock(codeBlock *ast.FencedCodeBlock) *MakeDoCodeBlock {
	n := &MakeDoCodeBlock{
		info:       codeBlock.Info,
		directives: make([]*Directive, 0, 2),
	}
	// Copy lines from original code block
	lines := codeBlock.Lines()
	n.SetLines(lines)
	return n
}

// OutputBlock returns the stdout output block associated with this code block, if any.
func (n *MakeDoCodeBlock) OutputBlock() *ast.FencedCodeBlock {
	return n.outputBlock
}

// SetOutputBlock sets the stdout output block associated with this code block.
func (n *MakeDoCodeBlock) SetOutputBlock(b *ast.FencedCodeBlock) {
	n.outputBlock = b
}

// Kind returns the NodeKind for MakeDoCodeBlock.
func (n *MakeDoCodeBlock) Kind() ast.NodeKind {
	return KindMakeDoCodeBlock
}

// Language returns the language specified in the fence info.
func (n *MakeDoCodeBlock) Language(source []byte) []byte {
	if n.info == nil {
		return nil
	}
	segment := n.info.Segment
	info := segment.Value(source)
	// Language is the first word of info
	for i, b := range info {
		if b == ' ' || b == '\t' {
			return info[:i]
		}
	}
	return info
}

// Info returns the full info string from the fence.
func (n *MakeDoCodeBlock) Info() *ast.Text {
	return n.info
}

// Code returns the code content as bytes.
func (n *MakeDoCodeBlock) Code(source []byte) []byte {
	lines := n.Lines()
	if lines.Len() == 0 {
		return nil
	}

	// Calculate total length for pre-allocation
	var totalLen int
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		totalLen += line.Len()
	}

	result := make([]byte, 0, totalLen)
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		result = append(result, line.Value(source)...)
	}
	return result
}

// Directives returns all directives associated with this code block.
func (n *MakeDoCodeBlock) Directives() []*Directive {
	return n.directives
}

// AddDirective adds a directive to this code block.
func (n *MakeDoCodeBlock) AddDirective(d *Directive) {
	n.directives = append(n.directives, d)
}

// HasDirective checks if a directive with the given kind exists.
func (n *MakeDoCodeBlock) HasDirective(kind DirectiveKind) bool {
	for _, d := range n.directives {
		if d.Kind == kind {
			return true
		}
	}
	return false
}

// GetDirective returns the first directive with the given kind.
func (n *MakeDoCodeBlock) GetDirective(kind DirectiveKind) *Directive {
	for _, d := range n.directives {
		if d.Kind == kind {
			return d
		}
	}
	return nil
}

// IsRaw returns true as code blocks contain raw content.
func (n *MakeDoCodeBlock) IsRaw() bool {
	return true
}

// Dump implements ast.Node.Dump for debugging.
func (n *MakeDoCodeBlock) Dump(source []byte, level int) {
	lang := n.Language(source)
	dirCount := len(n.directives)
	ast.DumpHelper(n, source, level, map[string]string{
		"Language":   string(lang),
		"Directives": fmt.Sprintf("%d", dirCount),
	}, nil)
}

// CodeSegments returns the line segments for the code content.
// This allows zero-copy iteration over code lines.
func (n *MakeDoCodeBlock) CodeSegments() *text.Segments {
	return n.Lines()
}
