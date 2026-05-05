package handlers

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestVerifyMarkdownRunsSetupBlockWithoutCountingTest(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	tmpDir := t.TempDir()
	markerFile := filepath.Join(tmpDir, "setup-ran.txt")
	mdPath := filepath.Join(tmpDir, "test.md")
	md := "```bash\n" +
		"echo ok > " + strconv.Quote(markerFile) + "\n" +
		"```\n"

	if err := os.WriteFile(mdPath, []byte(md), 0644); err != nil {
		t.Fatalf("failed to write markdown file: %v", err)
	}

	var verifyErr error
	output := captureStdout(t, func() {
		verifyErr = VerifyMarkdown(mdPath)
	})

	if verifyErr != nil {
		t.Fatalf("VerifyMarkdown() error = %v, want nil", verifyErr)
	}
	if !strings.Contains(output, "0/0 tests passed") {
		t.Fatalf("expected setup block to be excluded from test count, got output:\n%s", output)
	}

	markerContent, err := os.ReadFile(markerFile)
	if err != nil {
		t.Fatalf("expected setup marker file to be created: %v", err)
	}
	if strings.TrimSpace(string(markerContent)) != "ok" {
		t.Fatalf("marker file content = %q, want %q", string(markerContent), "ok")
	}
}

func TestVerifyMarkdownFailsOnSetupBlockNonZeroExit(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	tmpDir := t.TempDir()
	mdPath := filepath.Join(tmpDir, "test.md")
	md := "```bash\nexit 7\n```\n"

	if err := os.WriteFile(mdPath, []byte(md), 0644); err != nil {
		t.Fatalf("failed to write markdown file: %v", err)
	}

	var verifyErr error
	_ = captureStdout(t, func() {
		verifyErr = VerifyMarkdown(mdPath)
	})

	if verifyErr == nil {
		t.Fatal("VerifyMarkdown() error = nil, want non-nil")
	}
	if !strings.Contains(verifyErr.Error(), "setup block at line 2 failed: command exited with code 7") {
		t.Fatalf("unexpected error: %v", verifyErr)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	defer r.Close()

	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close stdout writer: %v", err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read captured stdout: %v", err)
	}
	return string(out)
}
