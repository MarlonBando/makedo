package nodes

import (
	"bytes"

	"github.com/yuin/goldmark/text"
)

var (
	commentStart = []byte("<!--")
	commentEnd   = []byte("-->")
)

// Directive represents a parsed comment directive (e.g., "out hello world").
// It stores text.Segment references for zero-copy access to source.
type Directive struct {
	Keyword text.Segment
	Content text.Segment
}

// KeywordString returns the keyword as a string.
// Only allocates when called.
func (d *Directive) KeywordString(source []byte) string {
	return string(d.Keyword.Value(source))
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
//   - Keyword is not registered in the registry
//
// Format: <!-- keyword content -->
func ParseDirective(data []byte, start int, registry *Registry) (*Directive, bool) {
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

	// Check if keyword is registered
	if !registry.IsValid(string(keyword)) {
		return nil, false
	}

	// Calculate segment offsets relative to start
	// start is the offset where 'data' begins in source
	prefixLen := len(commentStart)
	leadingSpaces := countLeadingSpaces(data[prefixLen:])

	keywordStart := start + prefixLen + leadingSpaces
	keywordSeg := text.NewSegment(keywordStart, keywordStart+len(keyword))

	var contentSeg text.Segment
	if content != nil {
		contentStart := keywordStart + len(keyword) + 1 + countLeadingSpaces(inner[keywordEnd+1:])
		contentSeg = text.NewSegment(contentStart, contentStart+len(content))
	}

	return &Directive{
		Keyword: keywordSeg,
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
