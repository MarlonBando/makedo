---
title: "exit"
weight: 7
---

The `exit` directive allows you to assert that a block of code finishes with a specific exit code.

By default, if a code block has **no directives at all**, MakeDo implicitly expects it to complete successfully with an exit code of `0`. However, if you add *any* directive to a block (such as `out`, `cmd`, or `exit`), MakeDo stops implicitly checking the exit code. The block's success is then determined entirely by whether the provided directives pass.

The `exit` directive is useful when you specifically want to assert the exit code—either to ensure it succeeds (`exit 0`) alongside other directives, or when testing negative scenarios where you *expect* a command to fail.

## Syntax
`<!-- exit [code] -->`
- `code`: Must be a single integer representing the expected exit code.

## Positive Assertion
Checks if the block finished exactly with the specified exit code.

```markdown
Run a command that we expect to fail with exit code 1:

` ` `bash
bash -c 'exit 1'
` ` `
<!-- exit 1 -->
```

## Negative Assertion (`!exit`)
The `exit` directive supports negation to assert that the block did **not** exit with a specific code.
This acts similarly to the regular `exit` directive by suppressing the default failure behavior, allowing any other exit code to pass.

```markdown
Ensure the command doesn't exit with code 0:

` ` `bash
bash -c 'exit 2'
` ` `
<!-- !exit 0 -->
```

## Important Notes
- **Single Directive:** Only one `exit` directive is allowed per test block. Multiple `exit` directives will result in a syntax error and fail the block immediately.
- **Background Processes:** The `exit` directive is only evaluated after the process finishes executing.
