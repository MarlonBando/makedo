---
title: "pwd"
weight: 3
bookToc: true
---

# The `pwd` Directive

The `pwd` directive asserts that the current working directory matches a given pattern. It supports a single `*` wildcard at the start or end of the pattern.

## Suffix Match

Match the end of the current directory path:

```bash
echo "checking pwd ends with makedo"
```
<!-- pwd makedo -->

## Prefix Match

Use a wildcard at the end to match the beginning of the path:

```bash
echo "checking pwd starts with /home"
```
<!-- pwd /home* -->

## Negation

Use `!pwd` to assert you are NOT in a specific directory:

```bash
echo "not in /etc"
```
<!-- !pwd /etc -->
