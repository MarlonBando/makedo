---
title: "Directives"
weight: 2
bookCollapseSection: true
bookToc: true
---

# Directives

Directives are HTML comments placed immediately after a fenced code block. They define assertions that MakeDo evaluates against the code block's execution.

## Syntax

The general format is:

```html
<!-- [!]keyword content -->
```
<!-- skip -->

- **`!` (Negation)**: An optional exclamation mark before the keyword negates the condition
- **`keyword`**: Must be one of the recognized directive types
- **`content`**: The string pattern or command to evaluate

## Multiple Directives

A single code block can have multiple directives. MakeDo executes the block once, then evaluates each directive in order:

```bash
mkdir -p /tmp/makedo_doc_multi
echo "multi-directive test"
```
<!-- out multi-directive -->
<!-- cmd test -d /tmp/makedo_doc_multi -->

```bash
rm -rf /tmp/makedo_doc_multi
```
<!-- cmd test ! -d /tmp/makedo_doc_multi -->

Note that only consecutive directive comments are collected. A regular HTML comment breaks the chain.

## Visible Directives

By default, HTML comments are invisible in rendered Markdown. If you want directives to be visible in your documentation, wrap them in backticks (inline code):

```bash
echo "visible directive"
```
`<!-- out visible -->`

## Available Directives

| Keyword | Description |
| :--- | :--- |
| `out` | Assert that stdout/stderr contains a literal string |
| `outr` | Assert that stdout/stderr matches a regular expression |
| `cmd` | Run a command; pass if it exits with code 0 |
| `pwd` | Assert the current working directory matches a pattern |
| `checkpath` | Assert a file or directory exists on the filesystem |
| `skip` | Tell MakeDo to ignore this block entirely |

Each directive is documented in detail in the following subsections.
