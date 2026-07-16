package handlers

import (
	"testing"
	"makedo/internal/engine"
)

func TestRunMarkdownFileMissingFile(t *testing.T) {
	ctx, err := engine.NewRunContext("")
	if err != nil {
		t.Fatalf("failed to create run context: %v", err)
	}
	defer ctx.Cleanup()
	err = RunMarkdownFile("/definitely/not/a/real/file.md", ctx)
	if err == nil {
		t.Fatalf("expected an error for a missing markdown file")
	}
}
