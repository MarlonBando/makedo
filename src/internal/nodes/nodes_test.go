package nodes

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	// Empty registry
	if r.IsValid("out") {
		t.Error("expected 'out' to be invalid in empty registry")
	}

	// Register keyword
	r.Register("out")
	if !r.IsValid("out") {
		t.Error("expected 'out' to be valid after registration")
	}
	if r.IsValid("skip") {
		t.Error("expected 'skip' to be invalid")
	}

	// Register multiple
	r.Register("skip")
	r.Register("config")
	keywords := r.Keywords()
	if len(keywords) != 3 {
		t.Errorf("expected 3 keywords, got %d", len(keywords))
	}
}

func TestParseDirective(t *testing.T) {
	r := NewRegistry()
	r.Register("out")
	r.Register("skip")

	tests := []struct {
		name    string
		input   string
		wantOK  bool
		keyword string
		content string
	}{
		{
			name:    "valid directive with content",
			input:   "<!-- out hello world -->",
			wantOK:  true,
			keyword: "out",
			content: "hello world",
		},
		{
			name:    "valid directive without content",
			input:   "<!-- skip -->",
			wantOK:  true,
			keyword: "skip",
			content: "",
		},
		{
			name:    "valid directive with extra whitespace",
			input:   "<!--   out   hello   -->",
			wantOK:  true,
			keyword: "out",
			content: "hello",
		},
		{
			name:   "unregistered keyword",
			input:  "<!-- config value -->",
			wantOK: false,
		},
		{
			name:   "regular comment",
			input:  "<!-- just a comment -->",
			wantOK: false,
		},
		{
			name:   "not a comment",
			input:  "<div>hello</div>",
			wantOK: false,
		},
		{
			name:   "empty comment",
			input:  "<!---->",
			wantOK: false,
		},
		{
			name:   "whitespace only comment",
			input:  "<!--   -->",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte(tc.input)
			d, ok := ParseDirective(source, 0, r)

			if ok != tc.wantOK {
				t.Errorf("ParseDirective() ok = %v, want %v", ok, tc.wantOK)
				return
			}

			if !ok {
				return
			}

			gotKeyword := d.KeywordString(source)
			if gotKeyword != tc.keyword {
				t.Errorf("keyword = %q, want %q", gotKeyword, tc.keyword)
			}

			gotContent := d.ContentString(source)
			if gotContent != tc.content {
				t.Errorf("content = %q, want %q", gotContent, tc.content)
			}
		})
	}
}

func TestMakeDoTransformer(t *testing.T) {
	tests := []struct {
		name          string
		markdown      string
		wantMakeDo    int // Number of MakeDoCodeBlock nodes expected
		wantFenced    int // Number of FencedCodeBlock nodes expected
		wantDirective string
	}{
		{
			name:          "code block with out directive",
			markdown:      "```bash\necho hello\n```\n<!-- out hello -->",
			wantMakeDo:    1,
			wantFenced:    0,
			wantDirective: "out",
		},
		{
			name:       "code block without directive",
			markdown:   "```bash\necho hello\n```",
			wantMakeDo: 0,
			wantFenced: 1,
		},
		{
			name:       "code block with regular comment",
			markdown:   "```bash\necho hello\n```\n<!-- just a note -->",
			wantMakeDo: 0,
			wantFenced: 1,
		},
		{
			name:       "code block with unregistered keyword",
			markdown:   "```bash\necho hello\n```\n<!-- config something -->",
			wantMakeDo: 0,
			wantFenced: 1,
		},
		{
			name:       "multiple code blocks mixed",
			markdown:   "```bash\necho one\n```\n<!-- out one -->\n\n```bash\necho two\n```",
			wantMakeDo: 1,
			wantFenced: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(NewMakeDoExtension()),
			)
			source := []byte(tc.markdown)
			reader := text.NewReader(source)
			doc := md.Parser().Parse(reader)

			var makeDoCount, fencedCount int
			var foundDirective string

			_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if !entering {
					return ast.WalkContinue, nil
				}
				switch n.Kind() {
				case KindMakeDoCodeBlock:
					makeDoCount++
					block := n.(*MakeDoCodeBlock)
					if len(block.Directives()) > 0 {
						foundDirective = block.Directives()[0].KeywordString(source)
					}
				case ast.KindFencedCodeBlock:
					fencedCount++
				}
				return ast.WalkContinue, nil
			})

			if makeDoCount != tc.wantMakeDo {
				t.Errorf("MakeDoCodeBlock count = %d, want %d", makeDoCount, tc.wantMakeDo)
			}
			if fencedCount != tc.wantFenced {
				t.Errorf("FencedCodeBlock count = %d, want %d", fencedCount, tc.wantFenced)
			}
			if tc.wantDirective != "" && foundDirective != tc.wantDirective {
				t.Errorf("directive = %q, want %q", foundDirective, tc.wantDirective)
			}
		})
	}
}

func TestMakeDoCodeBlock(t *testing.T) {
	markdown := "```bash\necho hello\necho world\n```\n<!-- out hello -->"

	md := goldmark.New(
		goldmark.WithExtensions(NewMakeDoExtension()),
	)
	source := []byte(markdown)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var block *MakeDoCodeBlock
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == KindMakeDoCodeBlock {
			block = n.(*MakeDoCodeBlock)
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})

	if block == nil {
		t.Fatal("expected MakeDoCodeBlock not found")
	}

	// Test Code()
	code := block.Code(source)
	expectedCode := "echo hello\necho world\n"
	if string(code) != expectedCode {
		t.Errorf("Code() = %q, want %q", string(code), expectedCode)
	}

	// Test Language()
	lang := block.Language(source)
	if string(lang) != "bash" {
		t.Errorf("Language() = %q, want %q", string(lang), "bash")
	}

	// Test Directives()
	directives := block.Directives()
	if len(directives) != 1 {
		t.Fatalf("Directives() count = %d, want 1", len(directives))
	}

	// Test HasDirective
	if !block.HasDirective("out", source) {
		t.Error("HasDirective('out') = false, want true")
	}
	if block.HasDirective("skip", source) {
		t.Error("HasDirective('skip') = true, want false")
	}

	// Test GetDirective
	d := block.GetDirective("out", source)
	if d == nil {
		t.Fatal("GetDirective('out') returned nil")
	}
	if d.ContentString(source) != "hello" {
		t.Errorf("directive content = %q, want %q", d.ContentString(source), "hello")
	}

	// Test IsRaw
	if !block.IsRaw() {
		t.Error("IsRaw() = false, want true")
	}
}

func TestNewMakeDoExtensionWithKeywords(t *testing.T) {
	ext := NewMakeDoExtensionWithKeywords("custom", "skip", "expect")

	if !ext.Registry.IsValid("custom") {
		t.Error("expected 'custom' to be valid")
	}
	if !ext.Registry.IsValid("skip") {
		t.Error("expected 'skip' to be valid")
	}
	if ext.Registry.IsValid("out") {
		t.Error("expected 'out' to be invalid (not registered)")
	}
}
