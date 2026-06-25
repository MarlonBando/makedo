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

	if execResult.Err != nil {
		outcome.Passed = false
		outcome.FailReason = execResult.Err
		return outcome
	}

	if execResult.Status == Completed && execResult.ExitCode > 0 {
		outcome.Passed = false
		outcome.FailReason = fmt.Errorf("command exited with code %d", execResult.ExitCode)
	} else if execResult.Status == Stalled && len(directives) > 0 {
		outcome.Passed = false
		outcome.FailReason = fmt.Errorf("command stalled before directives passed")
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

	if outcome.Passed && len(directives) > 0 {
		outcome.FinalOutput = SubstituteOutput(bytes.TrimSpace(execResult.Output), directives, source)
	} else {
		outcome.FinalOutput = bytes.TrimSpace(execResult.Output)
	}

	return outcome
}

// testDirective tests a single directive against execution result
func testDirective(d *nodes.Directive, execResult *Result, source []byte, startLine int, patterns map[*nodes.Directive]*regexp.Regexp, ctx *RunContext) *TestResult {
	// Handle non-zero exit for completed commands
	if execResult.Status == Completed && execResult.ExitCode != 0 {
		return &TestResult{
			Passed:    false,
			StartLine: startLine,
			Expected:  "command to succeed",
			Actual:    string(execResult.Output),
			Error:     fmt.Errorf("command exited with code %d", execResult.ExitCode),
		}
	}

	// If executor returned Ready, cmd directives already passed during execution
	if d.Kind == nodes.DirectiveCmd && execResult.Status == Ready {
		return &TestResult{
			Passed:    true,
			StartLine: startLine,
			Expected:  "exit 0",
			Actual:    d.ContentString(source),
		}
	}

	// Use shared directive checking
	check := CheckDirective(execResult.Output, d, source, patterns, ctx)
	return &TestResult{
		Passed:    check.Passed,
		StartLine: startLine,
		Expected:  check.Expected,
		Actual:    check.Actual,
		Error:     check.Err,
	}
}
