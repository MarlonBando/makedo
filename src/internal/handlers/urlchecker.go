package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"makedo/internal/engine"
	"makedo/internal/nodes"
	"makedo/internal/version"

	"github.com/yuin/goldmark/ast"
)

const (
	urlWorkers = 8
	urlTimeout = 10 * time.Second
)

// urlClient is shared so the default transport reuses connections.
var urlClient = &http.Client{Timeout: urlTimeout}

// urlUserAgent is built once at startup.
var urlUserAgent = "makedo/" + version.Get()

type urlJob struct {
	url       string
	startLine int
}

// CheckURLs walks doc, finds every http(s) Link/Image/AutoLink, GETs each one
// in a bounded pool, and returns the results as *engine.TestResult so the
// regular tester reporter can print them alongside directive tests.
func CheckURLs(doc ast.Node, source []byte) []*engine.TestResult {
	// TODO: introduce dedupe mechanism (map[string]struct{} of URLs, fan results back to occurrences).
	jobs := make([]urlJob, 0, 16)

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		dest, ok := nodes.IsLink(n, source)
		if !ok {
			return ast.WalkContinue, nil
		}
		jobs = append(jobs, urlJob{
			url:       string(dest),
			startLine: lineForNode(n, source),
		})
		return ast.WalkContinue, nil
	})

	if len(jobs) == 0 {
		return nil
	}

	results := make([]*engine.TestResult, len(jobs))
	sem := make(chan struct{}, urlWorkers)
	var wg sync.WaitGroup

	for i := range jobs {
		i := i
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = checkOne(jobs[i])
		}()
	}
	wg.Wait()
	return results
}

func checkOne(j urlJob) *engine.TestResult {
	expected := fmt.Sprintf("GET %s → HTTP 2xx", j.url)
	req, err := http.NewRequest(http.MethodGet, j.url, nil)
	if err != nil {
		return &engine.TestResult{
			Passed:    false,
			StartLine: j.startLine,
			Expected:  expected,
			Actual:    err.Error(),
			Error:     err,
		}
	}
	req.Header.Set("User-Agent", urlUserAgent)

	resp, err := urlClient.Do(req)
	if err != nil {
		return &engine.TestResult{
			Passed:    false,
			StartLine: j.startLine,
			Expected:  expected,
			Actual:    err.Error(),
			Error:     err,
		}
	}
	resp.Body.Close()

	passed := resp.StatusCode >= 200 && resp.StatusCode < 300
	return &engine.TestResult{
		Passed:    passed,
		StartLine: j.startLine,
		Expected:  expected,
		Actual:    fmt.Sprintf("HTTP %d", resp.StatusCode),
	}
}

// lineForNode returns the 1-indexed line of the enclosing block of n, or 0 if unknown.
// ponytail: enclosing-block granularity; refine to exact column if users ask.
func lineForNode(n ast.Node, source []byte) int {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if lines := p.Lines(); lines != nil && lines.Len() > 0 {
			return bytes.Count(source[:lines.At(0).Start], []byte{'\n'}) + 1
		}
	}
	return 0
}
