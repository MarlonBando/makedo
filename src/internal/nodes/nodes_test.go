package nodes

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestDirectiveKind(t *testing.T) {
	// Test ParseDirectiveKind
	tests := []struct {
		keyword []byte
		want    DirectiveKind
	}{
		{[]byte("out"), DirectiveOut},
		{[]byte("cmd"), DirectiveCmd},
		{[]byte("unknown"), DirectiveUnknown},
		{[]byte(""), DirectiveUnknown},
	}

	for _, tc := range tests {
		got := ParseDirectiveKind(tc.keyword)
		if got != tc.want {
			t.Errorf("ParseDirectiveKind(%q) = %v, want %v", tc.keyword, got, tc.want)
		}
	}

	// Test IsValidKeyword
	if !IsValidKeyword([]byte("out")) {
		t.Error("expected 'out' to be valid")
	}
	if !IsValidKeyword([]byte("cmd")) {
		t.Error("expected 'cmd' to be valid")
	}
	if IsValidKeyword([]byte("unknown")) {
		t.Error("expected 'unknown' to be invalid")
	}

	// Test String()
	if DirectiveOut.String() != "out" {
		t.Errorf("DirectiveOut.String() = %q, want %q", DirectiveOut.String(), "out")
	}
	if DirectiveCmd.String() != "cmd" {
		t.Errorf("DirectiveCmd.String() = %q, want %q", DirectiveCmd.String(), "cmd")
	}
	if DirectiveUnknown.String() != "unknown" {
		t.Errorf("DirectiveUnknown.String() = %q, want %q", DirectiveUnknown.String(), "unknown")
	}
}

func TestParseDirective(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOK  bool
		kind    DirectiveKind
		content string
	}{
		{
			name:    "valid directive with content",
			input:   "<!-- out hello world -->",
			wantOK:  true,
			kind:    DirectiveOut,
			content: "hello world",
		},
		{
			name:    "valid directive without content",
			input:   "<!-- out -->",
			wantOK:  true,
			kind:    DirectiveOut,
			content: "",
		},
		{
			name:    "valid cmd directive",
			input:   "<!-- cmd ls -la -->",
			wantOK:  true,
			kind:    DirectiveCmd,
			content: "ls -la",
		},
		{
			name:    "valid directive with extra whitespace",
			input:   "<!--   out   hello   -->",
			wantOK:  true,
			kind:    DirectiveOut,
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
		{
			name:    "negated out directive",
			input:   "<!-- !out failed -->",
			wantOK:  true,
			kind:    DirectiveOut,
			content: "failed",
		},
		{
			name:    "negated cmd directive",
			input:   "<!-- !cmd exit 1 -->",
			wantOK:  true,
			kind:    DirectiveCmd,
			content: "exit 1",
		},
		{
			name:    "checkpath directive",
			input:   "<!-- checkpath foo.txt -->",
			wantOK:  true,
			kind:    DirectiveCheckpath,
			content: "foo.txt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte(tc.input)
			d, ok := ParseDirective(source, 0)

			if ok != tc.wantOK {
				t.Errorf("ParseDirective() ok = %v, want %v", ok, tc.wantOK)
				return
			}

			if !ok {
				return
			}

			if d.Kind != tc.kind {
				t.Errorf("kind = %v, want %v", d.Kind, tc.kind)
			}

			// For our new negated tests, verify Negated flag is set properly
			if len(tc.name) > 7 && tc.name[:7] == "negated" {
				if !d.Negated {
					t.Errorf("expected directive to be Negated = true")
				}
			} else {
				if d.Negated {
					t.Errorf("expected directive to be Negated = false")
				}
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
		name       string
		markdown   string
		wantMakeDo int // Number of MakeDoCodeBlock nodes expected
		wantFenced int // Number of FencedCodeBlock nodes expected
		wantKind   DirectiveKind
	}{
		{
			name:       "code block with out directive",
			markdown:   "```bash\necho hello\n```\n<!-- out hello -->",
			wantMakeDo: 1,
			wantFenced: 0,
			wantKind:   DirectiveOut,
		},
		{
			name:       "code block with cmd directive",
			markdown:   "```bash\necho hello\n```\n<!-- cmd ls -->",
			wantMakeDo: 1,
			wantFenced: 0,
			wantKind:   DirectiveCmd,
		},
		{
			name:       "code block with checkpath directive",
			markdown:   "```bash\necho hello\n```\n<!-- checkpath foo.txt -->",
			wantMakeDo: 1,
			wantFenced: 0,
			wantKind:   DirectiveCheckpath,
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
			var foundKind DirectiveKind

			_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if !entering {
					return ast.WalkContinue, nil
				}
				switch n.Kind() {
				case KindMakeDoCodeBlock:
					makeDoCount++
					block := n.(*MakeDoCodeBlock)
					if len(block.Directives()) > 0 {
						foundKind = block.Directives()[0].Kind
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
			if tc.wantKind != DirectiveUnknown && foundKind != tc.wantKind {
				t.Errorf("directive kind = %v, want %v", foundKind, tc.wantKind)
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
	if !block.HasDirective(DirectiveOut) {
		t.Error("HasDirective(DirectiveOut) = false, want true")
	}
	if block.HasDirective(DirectiveCmd) {
		t.Error("HasDirective(DirectiveCmd) = true, want false")
	}

	// Test GetDirective
	d := block.GetDirective(DirectiveOut)
	if d == nil {
		t.Fatal("GetDirective(DirectiveOut) returned nil")
	}
	if d.ContentString(source) != "hello" {
		t.Errorf("directive content = %q, want %q", d.ContentString(source), "hello")
	}

	// Test IsRaw
	if !block.IsRaw() {
		t.Error("IsRaw() = false, want true")
	}
}
