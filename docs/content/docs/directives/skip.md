---
title: "skip"
weight: 5
bookToc: true
---

# The `skip` Directive

The `skip` directive tells MakeDo to completely ignore a code block. The block will not be executed, tested, or embedded. It remains as regular documentation-only content.

## Usage

Add `<!-- skip -->` after any code block to prevent MakeDo from processing it:

```bash
echo "This code block will NOT be executed by MakeDo"
```
<!-- skip -->

This is useful for:
- Showing example syntax that should not be run
- Documenting commands that require specific environments
- Including pseudo-code or template commands

## Skip Wins Over Other Directives

If `skip` appears alongside other directives, the block is still skipped entirely. The `skip` directive always takes priority.

## Visible Skip

You can wrap the skip directive in backticks to make it visible in rendered Markdown:

```bash
echo "This is also skipped"
```
`<!-- skip -->`
