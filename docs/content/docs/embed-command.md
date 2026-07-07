---
title: "Embed Command"
weight: 6
bookToc: true
---

# The `embed` Command

The `makedo embed` command executes code blocks and injects their output back into the Markdown file. This is the documentation generation workflow.

## How It Works

1. MakeDo executes blocks using the same model as `test`
2. Upon successful execution, it captures the stdout/stderr output
3. It injects or replaces a fenced code block marked as ` ```stdout ` below the directives
4. If a block fails its directives, `embed` skips that block and warns the user

## Example

Given a Markdown file containing:

````markdown
```bash
echo "hello embed"
```
<!-- out hello embed -->
````
<!-- skip -->

Running `makedo embed` will produce:

````markdown
```bash
echo "hello embed"
```
<!-- out hello embed -->

```stdout
hello embed
```
````
<!-- skip -->

## Dynamic Type Substitution

During embed, if dynamic type placeholders were used in directives, MakeDo replaces actual runtime values back into placeholder syntax. This ensures the embedded output remains deterministic.

For example, an actual IP like `192.168.1.1` would be replaced with `${{ip}}` in the embedded output.

## test vs embed

| Feature | `test` | `embed` |
| :--- | :--- | :--- |
| Executes code blocks | ✓ | ✓ |
| Verifies directives | ✓ | ✓ |
| Modifies the file | ✗ | ✓ |
| Injects stdout blocks | ✗ | ✓ |
| Dynamic type substitution | ✗ | ✓ |
