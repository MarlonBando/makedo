---
name: makedo-agent-skill
description: Learn how to use makedo for markdown-based testing and embedding
license: MIT
metadata:
  audience: developers, coding agents
  workflow: testing, documentation
---

## What is makedo?

Makedo is a tool that brings integration testing directly inside your markdown documentation. Documentation becomes the tests, and tests become documentation.

Makedo allows you to:
- Verify that your documentation is still saying the truth
- Run integration testing on your cli tool using markdown
- Run all the codes in a markdown file (useful for setup)

**Note for the agent**: This skill file is already an example of `makedo` usage! The code blocks below are followed by HTML comments that `makedo` uses to validate the output. You can run `makedo test skills/SKILL.md` to verify this file.

## When to use it

Use this tool when you need to:
- Write documentation that needs to be automatically tested against regressions.
- Embed CLI integration tests directly within your markdown files.
- Ensure that the commands in your README or docs actually produce the output they claim to produce.
- Use dynamic types like `${{date}}` or `${{uuid}}` to verify outputs that change on every run.

## Main command (test, embed, help)

Makedo works by looking for fenced code blocks followed by an HTML comment in this format: `<!-- directive content -->`.

### Help

You can check the main help command to see what makedo can do:

```bash
makedo -h
```
<!-- out A markdown-based task runner -->

### Test

The `test` command runs code blocks with "out" directives and verifies their output matches the expected pattern.

```bash
makedo test -h
```
<!-- out directive <content> -->

### Embed

The `embed` command runs fenced code blocks and embeds their output directly into the markdown file.

```bash
makedo embed -h
```
<!-- out Run fenced code blocks and embed output into the markdown file -->
