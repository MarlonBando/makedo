package handlers

import "testing"

func TestRunMarkdownFileMissingFile(t *testing.T) {
	err := RunMarkdownFile("/definitely/not/a/real/file.md")
	if err == nil {
		t.Fatalf("expected an error for a missing markdown file")
	}
}
