---
title: "out & outr"
weight: 1
bookToc: true
---

# The `out` and `outr` Directives

The `out` directive verifies that the output of a code block contains a given string. The `outr` directive matches using a regular expression.

## Basic Usage (out)

The simplest form checks if the output contains a substring:

```bash
echo "Hello World"
```
<!-- out Hello -->

```bash
ls /
```
<!-- out home -->

## Regex Matching (outr)

Use `outr` when you need regular expression patterns:

```bash
echo "exact"
```
<!-- outr ^exact$ -->

```bash
echo "file123.txt"
```
<!-- outr file[0-9]+\.txt -->

## Case Sensitivity

Patterns are case-sensitive by default:

```bash
echo "Linux"
```
<!-- out Linux -->

## Quantifiers

Common regex quantifiers work with `outr`: `*` (zero or more), `+` (one or more), `?` (zero or one):

```bash
echo "goooal"
```
<!-- outr go+al -->

```bash
echo "color"
```
<!-- outr colou?r -->

## Alternation

Use `|` for OR patterns:

```bash
echo "cat"
```
<!-- outr (cat|dog) -->

## Multiline Output

```bash
printf "line1\nline2\nline3"
```
<!-- outr line1\nline2\nline3 -->

## Empty Output

Match empty output with `^$`:

```bash
true
```
<!-- outr ^$ -->

## Negation

Use `!out` or `!outr` to assert the output does **not** contain a pattern:

```bash
echo "hello"
```
<!-- !out world -->
<!-- out hello -->

```bash
echo "123"
```
<!-- !outr [a-z]+ -->
