# The `out` Directive

The `out` directive verifies the output of a code block using regex pattern matching.
Add after a code block a html comment in this format `<!-- out ${regex_str} -->`

## Basic Substring Matching

By default, the pattern matches if it appears anywhere in the output.

```bash
echo "Hello world!"
```
<!-- out world -->

```bash
ls /
```
<!-- out home -->

## Exact Matching

Use regex anchors `^` (start) and `$` (end) for exact matches.

```bash
echo "exact"
```
<!-- out ^exact$ -->

```bash
echo "Hello world!"
```
<!-- out ^Hello world!$ -->

## Case Sensitivity

Patterns are case-sensitive by default.

```bash
echo "Linux"
```
<!-- out Linux -->

## Special Characters

Regex special characters must be escaped with backslash.

```bash
echo "test.log"
```
<!-- out test\.log -->

```bash
echo 'cost: $10'
```
<!-- out \$10 -->

## Character Classes

Use `[...]` to match any character in the set.

```bash
echo "file123.txt"
```
<!-- out file[0-9]+\.txt -->

```bash
echo "hello"
```
<!-- out h[aeiou]llo -->

## Quantifiers

Common quantifiers: `*` (zero or more), `+` (one or more), `?` (zero or one).

```bash
echo "goooal"
```
<!-- out go+al -->

```bash
echo "color"
```
<!-- out colou?r -->

## Alternation

Use `|` for OR patterns.

```bash
echo "cat"
```
<!-- out (cat|dog) -->

## Real-World Example: Version Checking

Check if git version is 2.x or higher.

```bash
git --version
```
<!-- out ^git version [2-9]\. -->

## Multiline Output

By default, newlines are trimmed from the end. Use `(?s)` flag or match newlines explicitly.

```bash
printf "line1\nline2\nline3"
```
<!-- out line1\nline2\nline3 -->

## Empty Output

Match empty output with `^$`.

```bash
true
```
<!-- out ^$ -->

## Whitespace Matching

Use `\s` for whitespace, `\t` for tabs.

```bash
echo "hello   world"
```
<!-- out hello\s+world -->
