package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"makedo/internal/nodes"
)

type DirectiveResult struct {
	Passed   bool
	Expected string
	Actual   string
	Err      error
}

// CheckDirective evaluates a directive against output and returns the result.
// TODO: The `output` parameter currently represents the combined stdout+stderr stream.
// Future implementation of an `err` directive will require splitting this signature
// to accept separate streams, or using a structured chunking mechanism.
func CheckDirective(output []byte, exitCode *int, d *nodes.Directive, source []byte, compiledPatterns map[*nodes.Directive]*regexp.Regexp, ctx *RunContext) *DirectiveResult {
	content := d.ContentString(source)

	var res *DirectiveResult
	switch d.Kind {
	case nodes.DirectiveExit:
		res = checkExit(exitCode, content)
	case nodes.DirectiveOut:
		if re, ok := compiledPatterns[d]; ok {
			res = checkOutPrecompiled(output, re, content)
		} else {
			res = checkOut(output, content)
		}
	case nodes.DirectiveOutRegex:
		if re, ok := compiledPatterns[d]; ok {
			res = checkOutPrecompiled(output, re, content)
		} else {
			res = checkOutRegex(output, content)
		}
	case nodes.DirectiveCmd:
		res = checkCmd(content, ctx)
	case nodes.DirectivePwd:
		res = checkPwd(content, ctx)
	case nodes.DirectiveCheckpath:
		res = checkPath(content, ctx)
	default:
		res = &DirectiveResult{Passed: true}
	}

	if d.Negated {
		// Do not invert structural errors
		isStructuralErr := res.Err != nil
		if isStructuralErr {
			// Check if it's an ExitError (which represents a command failure, not structural)
			if _, isExitError := res.Err.(*exec.ExitError); isExitError {
				isStructuralErr = false
			}
		}

		if !isStructuralErr {
			res.Passed = !res.Passed
			res.Expected = "NOT " + res.Expected
			if res.Passed {
				res.Err = nil
			}
		}
	}

	return res
}

func checkExit(exitCode *int, content string) *DirectiveResult {
	expectedExit, err := strconv.Atoi(strings.TrimSpace(content))
	if err != nil {
		return &DirectiveResult{
			Passed:   false,
			Expected: "integer exit code",
			Actual:   content,
			Err:      fmt.Errorf("invalid exit directive: must be an integer"),
		}
	}

	if exitCode == nil {
		// Process is still running, so it hasn't exited yet.
		return &DirectiveResult{
			Passed:   false,
			Expected: fmt.Sprintf("exit code %d", expectedExit),
			Actual:   "process still running",
		}
	}

	return &DirectiveResult{
		Passed:   *exitCode == expectedExit,
		Expected: fmt.Sprintf("exit code %d", expectedExit),
		Actual:   fmt.Sprintf("exit code %d", *exitCode),
	}
}

func checkOutPrecompiled(output []byte, re *regexp.Regexp, expected string) *DirectiveResult {
	actual := strings.TrimRight(string(output), "\n")
	return &DirectiveResult{
		Passed:   re.MatchString(actual),
		Expected: expected,
		Actual:   actual,
	}
}

func checkOut(output []byte, expected string) *DirectiveResult {
	expected = strings.TrimSpace(expected)
	actual := strings.TrimRight(string(output), "\n")
	return &DirectiveResult{
		Passed:   strings.Contains(actual, expected),
		Expected: expected,
		Actual:   actual,
	}
}

func checkOutRegex(output []byte, pattern string) *DirectiveResult {
	actual := strings.TrimRight(string(output), "\n")
	re, err := regexp.Compile(pattern)
	if err != nil {
		return &DirectiveResult{
			Passed:   false,
			Expected: pattern,
			Actual:   actual,
			Err:      fmt.Errorf("invalid regex '%s': %w", pattern, err),
		}
	}
	return &DirectiveResult{
		Passed:   re.MatchString(actual),
		Expected: pattern,
		Actual:   actual,
	}
}

func checkCmd(command string, ctx *RunContext) *DirectiveResult {
	cmdStr := command
	if ctx.EnvFile != "" {
		cmdStr = fmt.Sprintf("source %q\n%s", ctx.EnvFile, command)
	}
	shell, flag := getShell()
	cmd := exec.Command(shell, flag, cmdStr)
	if ctx.Cwd != "" {
		cmd.Dir = ctx.Cwd
	}
	err := cmd.Run()
	return &DirectiveResult{
		Passed:   err == nil,
		Expected: "exit 0",
		Actual:   command,
		Err:      err,
	}
}

func checkPwd(pattern string, ctx *RunContext) *DirectiveResult {
	pattern = strings.TrimSpace(pattern)
	cwd := ctx.Cwd
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return &DirectiveResult{
				Passed:   false,
				Expected: pattern,
				Err:      fmt.Errorf("failed to get working directory: %w", err),
			}
		}
	}

	normalizedCwd := normalizePath(cwd)
	normalizedPattern := normalizePath(pattern)

	matched, err := matchGlob(normalizedCwd, normalizedPattern)
	if err != nil {
		return &DirectiveResult{
			Passed:   false,
			Expected: pattern,
			Err:      err,
		}
	}
	return &DirectiveResult{
		Passed:   matched,
		Expected: pattern,
		Actual:   cwd,
	}
}

func checkPath(path string, ctx *RunContext) *DirectiveResult {
	path = strings.TrimSpace(path)
	// TODO: expand environment variables

	checkP := path
	if !filepath.IsAbs(checkP) && ctx.Cwd != "" {
		checkP = filepath.Join(ctx.Cwd, checkP)
	}
	_, err := os.Stat(checkP)
	return &DirectiveResult{
		Passed:   err == nil,
		Expected: path,
		Actual:   path,
	}
}

func matchGlob(path, pattern string) (bool, error) {
	starCount := strings.Count(pattern, "*")

	if starCount > 1 {
		return false, fmt.Errorf("pwd directive only allows a single wildcard (*) at the start or end of the pattern, got: %q", pattern)
	}

	if starCount == 1 {
		if strings.HasPrefix(pattern, "*") {
			return strings.HasSuffix(path, pattern[1:]), nil
		}
		if strings.HasSuffix(pattern, "*") {
			return strings.HasPrefix(path, pattern[:len(pattern)-1]), nil
		}
		return false, fmt.Errorf("pwd directive only allows a wildcard (*) at the start or end of the pattern, got: %q", pattern)
	}

	return strings.HasSuffix(path, pattern), nil
}

func normalizePath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	return p
}

// checkFastDirectives tests output-dependent directives (out, outr, pwd).
// These are cheap operations - no process spawn.
func checkFastDirectives(output []byte, directives []*nodes.Directive, source []byte, compiledPatterns map[*nodes.Directive]*regexp.Regexp, ctx *RunContext) bool {
	for _, d := range directives {
		switch d.Kind {
		case nodes.DirectiveOut, nodes.DirectiveOutRegex, nodes.DirectivePwd, nodes.DirectiveCheckpath:
			if !CheckDirective(output, nil, d, source, compiledPatterns, ctx).Passed {
				return false
			}
		}
	}
	return true
}

// checkAllDirectives tests all directives including cmd.
// Called on ticker to avoid spawning processes too frequently.
func checkAllDirectives(output []byte, directives []*nodes.Directive, source []byte, compiledPatterns map[*nodes.Directive]*regexp.Regexp, ctx *RunContext) bool {
	for _, d := range directives {
		if !CheckDirective(output, nil, d, source, compiledPatterns, ctx).Passed {
			return false
		}
	}
	return true
}

// hasCmdDirective checks if any directive is a cmd directive.
func hasCmdDirective(directives []*nodes.Directive) bool {
	for _, d := range directives {
		if d.Kind == nodes.DirectiveCmd {
			return true
		}
	}
	return false
}
