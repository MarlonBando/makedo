---
name: "makedo-test-generator"
description: "Generates embedded integration tests in Markdown documentation using the makedo tool syntax."
---

# Skill: Makedo Test Generator

## Context
You are an expert at writing integration tests using `makedo`. `makedo` executes code directly from Markdown fenced code blocks and validates the output or state using hidden HTML comment directives placed immediately after the block. Your goal is to turn documentation into verifiable integration tests.

## Syntax & Directives
Tests are structured by pairing a standard Markdown fenced code block with an HTML comment directive. 

**Format:**

```[language]
command to run
```
<!-- directive content -->

**Available Directives:**
* `out [content]`: Asserts that the standard output (stdout) contains the exact `[content]`.
* `outr [regex]`: Asserts that the standard output (stdout) matches the provided `[regex]`.
* `cmd [content]`: Runs a follow-up shell command to validate the state. If the command succeeds (exit code 0), the test passes.
* `pwd [content]`: Asserts that the current working directory path matches `[content]`.

## Strict Rules
1.  **Placement:** The validation HTML comment MUST be placed immediately after the closing backticks of the fenced code block.
2.  **Multiple Assertions:** You can use multiple directives for a single code block by stacking the HTML comments sequentially.
3.  **No Background Processes:** Long-lasting commands (like starting a web server or database) survive throughout the execution automatically. **DO NOT** use the background command operator (`&`). Write the command synchronously.
4.  **Visibility:** Do not write any explanations inside the HTML comment other than the strict `[directive] [content]` syntax, as `makedo` parses these exactly.

## Examples

### Example 1: Basic Output Validation
Checking if a help flag prints the expected output.

```bash
./bin/makedo -h
```
<!-- out makedo -->

### Example 2: State Validation via Follow-up Command
Running a command and then verifying the system state by executing another command.

```bash
echo "Pretending to install git..."
```
<!-- cmd git --version -->

### Example 3: Multiple Directives
Using multiple directives to validate both the current directory and the output.

```bash
mkdir -p /tmp/makedo-test && cd /tmp/makedo-test
echo "hello world"
```
<!-- pwd /tmp/makedo-test -->
<!-- out hello world -->

### Example 4: Long-lasting Commands
Starting a server without using the `&` background operator, then validating it with a regex pattern or even better with a curl command.

```bash
python -m http.server 8080
```
<!-- outr Serving HTTP on .* port 8080 -->
