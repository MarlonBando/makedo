---
title: "MakeDo"
type: docs
---

# MakeDo

**Documentation that tests itself.**

MakeDo is a tool that brings testing directly inside your Markdown documentation. Your documentation becomes the tests, and your tests become documentation.

With MakeDo you can:

- **Verify** that your documentation is still telling the truth
- **Run** integration and system tests on your CLI tools using Markdown
- **Embed** the output of commands automatically inside your docs
- **Execute** all the fenced code blocks in a Markdown file (useful for setup)

## Quick Start

Install makedo, then test any markdown file:

```bash
makedo test your-documentation.md
```
<!-- skip -->

Or embed command output back into your documentation:

```bash
makedo embed your-documentation.md
```
<!-- skip -->

## How It Works

MakeDo looks for fenced code blocks followed by HTML comment directives in the format `<!-- keyword content -->`. The content of the comment tells MakeDo what to expect from the code block output.

When the Markdown is rendered, MakeDo directives are hidden since they are HTML comments. If you want to make them visible in rendered documentation, wrap them in backticks.

## Example

Here is a simple example. We run `echo` and verify the output contains "Hello":

```bash
echo "Hello World"
```
`<!-- out Hello -->`

The directive `<!-- out Hello -->` tells MakeDo to verify that the output of the code block contains the string "Hello".
