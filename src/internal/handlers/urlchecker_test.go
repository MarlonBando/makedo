package handlers

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"makedo/internal/engine"
)

func newURLTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusFound)
	})
	mux.HandleFunc("/boom", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	return httptest.NewServer(mux)
}

func writeMD(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func runVerify(t *testing.T, mdPath string, checkURLs bool) (string, error) {
	t.Helper()
	var err error
	out := captureStdout(t, func() {
		ctx, cerr := engine.NewRunContext()
		if cerr != nil {
			t.Fatal(cerr)
		}
		defer ctx.Cleanup()
		ctx.CheckURLs = checkURLs
		err = VerifyMarkdown(mdPath, ctx)
	})
	return out, err
}

func TestCheckURLs_PassAndFail(t *testing.T) {
	srv := newURLTestServer(t)
	defer srv.Close()

	md := "Good: [a](" + srv.URL + "/ok)\n\nBad: [b](" + srv.URL + "/missing)\n"
	p := writeMD(t, md)

	out, err := runVerify(t, p, true)
	if err == nil {
		t.Fatalf("expected error from failing URL, got nil. out:\n%s", out)
	}
	if !strings.Contains(out, "1/2 tests passed") {
		t.Fatalf("want 1/2 passed, got:\n%s", out)
	}
	if !strings.Contains(out, "HTTP 404") {
		t.Fatalf("expected HTTP 404 in output:\n%s", out)
	}
	if !strings.Contains(out, "2 URLs checked") {
		t.Fatalf("expected '2 URLs checked' summary line:\n%s", out)
	}
}

func TestCheckURLs_RedirectFollowed(t *testing.T) {
	srv := newURLTestServer(t)
	defer srv.Close()

	md := "[r](" + srv.URL + "/redirect)\n"
	p := writeMD(t, md)

	out, err := runVerify(t, p, true)
	if err != nil {
		t.Fatalf("expected redirect to be followed to 200, got err %v. out:\n%s", err, out)
	}
	if !strings.Contains(out, "1/1 tests passed") {
		t.Fatalf("want 1/1 passed:\n%s", out)
	}
}

func TestCheckURLs_ImageAndAutoLink(t *testing.T) {
	srv := newURLTestServer(t)
	defer srv.Close()

	md := "![img](" + srv.URL + "/ok)\n\nauto: <" + srv.URL + "/ok>\n"
	p := writeMD(t, md)

	out, err := runVerify(t, p, true)
	if err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
	if !strings.Contains(out, "2/2 tests passed") {
		t.Fatalf("want 2/2 (image + autolink):\n%s", out)
	}
}

func TestCheckURLs_NonHTTPIgnored(t *testing.T) {
	md := "[mail](mailto:foo@example.com) [rel](./other.md) [anchor](#x)\n"
	p := writeMD(t, md)

	out, err := runVerify(t, p, true)
	if err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
	if !strings.Contains(out, "0 URLs checked") {
		t.Fatalf("want 0 URLs checked:\n%s", out)
	}
}

func TestCheckURLs_FlagOffSkipsChecks(t *testing.T) {
	// Point at a bogus URL; if the checker ran, it would fail.
	md := "[bad](http://127.0.0.1:1/nope)\n"
	p := writeMD(t, md)

	out, err := runVerify(t, p, false)
	if err != nil {
		t.Fatalf("flag off must not run URL checks; got err %v\n%s", err, out)
	}
	if strings.Contains(out, "URLs checked") {
		t.Fatalf("did not expect URL summary line when flag off:\n%s", out)
	}
}

func TestCheckURLs_ConnectionRefused(t *testing.T) {
	// Reserve a port then close it so connect fails fast.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("cannot reserve a port")
	}
	addr := ln.Addr().String()
	ln.Close()

	md := "[dead](http://" + addr + "/x)\n"
	p := writeMD(t, md)

	out, err := runVerify(t, p, true)
	if err == nil {
		t.Fatalf("expected error from connection refused\n%s", out)
	}
	if !strings.Contains(out, "0/1 tests passed") {
		t.Fatalf("want 0/1:\n%s", out)
	}
}
