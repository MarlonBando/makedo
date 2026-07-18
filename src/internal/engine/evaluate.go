package engine

import (
	"bytes"
	"fmt"
	"regexp"

	"makedo/internal/nodes"
)

type TestResult struct {
	Passed    bool
	StartLine int
	Expected  string
	Actual    string
	Error     error
}

type BlockOutcome struct {
	ExecResult  *Result       // Raw execution stream and status
	TestResults []*TestResult // Individual directive evaluations
	Passed      bool          // True if the block is considered successful (v = True)
	FailReason  error         // Reason for failure if Passed is false
	FinalOutput []byte        // Substituted output for embed
}

// EvaluateBlock executes the block, evaluates the directives, manages the process registry,
// and returns the unified outcome and substituted output.
func EvaluateBlock(code string, directives []*nodes.Directive, source []byte, lineNum int, ctx *RunContext) *BlockOutcome {
	outcome := &BlockOutcome{
		Passed: true,
	}

	patterns, err := PrecompileDirectives(directives, source)
	if err != nil {
		outcome.Passed = false
		outcome.FailReason = fmt.Errorf("failed to precompile directives: %w", err)
		outcome.ExecResult = &Result{Err: outcome.FailReason, ExitCode: -1}
		return outcome
	}

	execResult := Execute(ctx, code, directives, source, false)
	outcome.ExecResult = execResult

	if execResult.Process != nil && execResult.Status != Completed {
		ctx.Registry.Add(execResult.Process)
	}

	hasExitDirective := false
	for _, d := range directives {
		if d.Kind == nodes.DirectiveExit {
			if hasExitDirective {
				outcome.Passed = false
				outcome.FailReason = fmt.Errorf("multiple exit directives are not allowed in a single block")
				return outcome
			}
			hasExitDirective = true
		}
	}

	if execResult.Err != nil {
		outcome.Passed = false
		outcome.FailReason = fmt.Errorf("block at line %d crashed: %w", lineNum, execResult.Err)
		return outcome
	}

	hasAnyDirectives := len(directives) > 0

	if !hasAnyDirectives && execResult.Status == Completed && execResult.ExitCode > 0 {
		outcome.Passed = false
		outcome.FailReason = fmt.Errorf("command exited with code %d", execResult.ExitCode)
	}

	for _, d := range directives {
		testRes := testDirective(d, execResult, source, lineNum, patterns, ctx)
		if testRes == nil {
			continue
		}

		outcome.TestResults = append(outcome.TestResults, testRes)
		if testRes.Passed {
			continue
		}

		outcome.Passed = false
		if outcome.FailReason == nil { // Don't overwrite higher-level failures if already set
			if testRes.Error != nil {
				outcome.FailReason = testRes.Error
			} else {
				outcome.FailReason = fmt.Errorf("directive '%s' failed to match", string(d.Kind))
			}
		}
	}

	// We MUST use execResult.CleanOut (not Output) for FinalOutput.
	// Commands run inside a PTY inject `\r` (carriage returns) into the raw output buffer.
	// If we use the raw Output, `makedo embed` will permanently inject those hidden `\r`
	// characters directly into the markdown file, breaking regex matches down the line.
	if outcome.Passed && len(directives) > 0 {
		outcome.FinalOutput = SubstituteOutput(bytes.TrimSpace(execResult.CleanOut), directives, source)
	} else {
		outcome.FinalOutput = bytes.TrimSpace(execResult.CleanOut)
	}

	return outcome
}

// testDirective tests a single directive against execution result
func testDirective(d *nodes.Directive, execResult *Result, source []byte, startLine int, patterns map[*nodes.Directive]*regexp.Regexp, ctx *RunContext) *TestResult {

	// If executor returned Ready, cmd directives already passed during execution
	if d.Kind == nodes.DirectiveCmd && execResult.Status == Ready {
		return &TestResult{
			Passed:    true,
			StartLine: startLine,
			Expected:  "exit 0",
			Actual:    d.ContentString(source),
		}
	}

	// Use shared directive checking against the stripped output
	check := CheckDirective(execResult.CleanOut, &execResult.ExitCode, d, source, patterns, ctx)
	return &TestResult{
		Passed:    check.Passed,
		StartLine: startLine,
		Expected:  check.Expected,
		Actual:    check.Actual,
		Error:     check.Err,
	}
}
