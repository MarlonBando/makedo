package nodes

import (
	"bytes"

	"github.com/yuin/goldmark/text"
)

var (
	commentStart = []byte("<!--")
	commentEnd   = []byte("-->")
)

// Directive represents a parsed comment directive (e.g., "<!-- out hello world -->").
// It stores the resolved DirectiveKind and a text.Segment reference for zero-copy
// access to the content.
type Directive struct {
	Kind    DirectiveKind // The resolved directive type
	Content text.Segment  // Byte range for content in source (zero-copy)
}

// ContentString returns the content as a string.
// Only allocates when called.
func (d *Directive) ContentString(source []byte) string {
	return string(d.Content.Value(source))
}

// ContentBytes returns the content as bytes (zero-copy).
func (d *Directive) ContentBytes(source []byte) []byte {
	return d.Content.Value(source)
}

// ParseDirective parses an HTML comment into a Directive.
// Returns nil, false if:
//   - Not a valid HTML comment (<!-- ... -->)
//   - Keyword is not a recognized directive
//
// Format: <!-- keyword content -->
func ParseDirective(data []byte, start int) (*Directive, bool) {
	data = bytes.TrimSpace(data)

	// Check comment markers
	if !bytes.HasPrefix(data, commentStart) {
		return nil, false
	}
	if !bytes.HasSuffix(data, commentEnd) {
		return nil, false
	}

	// Extract inner content (between <!-- and -->)
	inner := data[len(commentStart) : len(data)-len(commentEnd)]
	inner = bytes.TrimSpace(inner)
	if len(inner) == 0 {
		return nil, false
	}

	// Find keyword (first word)
	keywordEnd := bytes.IndexByte(inner, ' ')
	var keyword, content []byte
	if keywordEnd == -1 {
		// Keyword only, no content
		keyword = inner
		content = nil
	} else {
		keyword = inner[:keywordEnd]
		content = bytes.TrimSpace(inner[keywordEnd+1:])
	}

	kind := ParseDirectiveKind(keyword)
	if kind == DirectiveUnknown {
		return nil, false
	}

	// Calculate content segment offset relative to start
	var contentSeg text.Segment
	if content != nil {
		prefixLen := len(commentStart)
		leadingSpaces := countLeadingSpaces(data[prefixLen:])
		keywordStart := prefixLen + leadingSpaces
		contentStart := start + keywordStart + len(keyword) + 1 + countLeadingSpaces(inner[keywordEnd+1:])
		contentSeg = text.NewSegment(contentStart, contentStart+len(content))
	}

	return &Directive{
		Kind:    kind,
		Content: contentSeg,
	}, true
}

func countLeadingSpaces(data []byte) int {
	for i, b := range data {
		if b != ' ' && b != '\t' {
			return i
		}
	}
	return len(data)
}
