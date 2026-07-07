---
title: "Getting Started"
weight: 1
bookToc: true
---

# Getting Started

MakeDo turns your Markdown documentation into executable test suites. Every fenced code block with a shell language identifier (`bash`, `sh`, `zsh`, or `shell`) is automatically detected and can be tested.

## Installation

Download the latest binary from the releases page or build it from source and place it in your `PATH`:

You can check if it's installed by checking the version with 
```bash
makedo -v
```
<!-- out ${{version}} -->

## Available Commands

MakeDo provides three main commands:

| Command | Description |
| :--- | :--- |
| `test` | Run code blocks and verify directives pass |
| `embed` | Run code blocks and inject output into the Markdown |
| `run` | Simply execute all code blocks |

You can see all available commands with:

```bash
makedo -h
```
<!-- out Available Commands -->
<!-- out test -->
<!-- out embed -->
<!-- out run -->

## Your First Test

Create a Markdown file called `hello.md` with a code block and a directive.


You can do it by copy and past the content of this block into the file
````markdown
```bash
echo hello world!
```
<!-- out hello world! -->
````

or by running this command in your terminal
```bash
printf '```bash\necho hello world!\n```\n<!-- out hello world! -->\n' > hello.md
```
<!-- checkpath hello.md -->

Then run:

```bash
makedo test hello.md
```
<!-- out 1/1 -->

MakeDo will execute the code block and verify the output matches the directive.

## Available Commands

MakeDo provides three main commands:

| Command | Description |
| :--- | :--- |
| `test` | Run code blocks and verify directives pass |
| `embed` | Run code blocks and inject output into the Markdown |
| `run` | Simply execute all code blocks |

You can see all available commands with:

```bash
makedo -h
```
<!-- out Available Commands -->
<!-- out test -->
<!-- out embed -->
<!-- out run -->
