---
title: "MakeDo"
type: docs
---

# MakeDo

**Test your CLI with Markdown.**

Write your CLI's tests in Markdown. The docs are the tests, and the tests are docs anyone can read.

Add a directive below a shell code block and MakeDo checks it:

```bash
echo "Hello World"
```
`<!-- out Hello -->`

The directive `<!-- out Hello -->` tells MakeDo to verify that the output of the code block contains the string "Hello".

## Why MakeDo

- **Docs that can't lie.** When an example stops working, the test fails.
- **CLI tests without a test framework.** No harness to write, just commands and the output you expect, in plain Markdown.
- **A real shell session.** Variables, `cd` and files carry over between blocks. Servers started with `&` keep running in the background and are cleaned up at the end.
- **Output that writes itself.** `makedo embed` runs your blocks and writes their real output into the doc.

## Quick Start

Test any Markdown file:

```bash
makedo test your-documentation.md
```
`<!-- skip -->`

Or embed command output back into it:

```bash
makedo embed your-documentation.md
```
`<!-- skip -->`

See [Getting Started]({{< relref "/docs/getting-started" >}}) to install MakeDo and write your first test.

## How It Works

MakeDo looks for fenced code blocks followed by HTML comment directives in the format `<!-- keyword content -->`. The content of the comment tells MakeDo what to expect from the code block output.

Because directives are HTML comments, a bare directive is hidden when the Markdown is rendered. Wrap it in backticks, as in the examples on this site, to make it visible.

## Tested with MakeDo

MakeDo tests itself: its test suite is written in Markdown, and every page of this documentation asserts its own examples.
