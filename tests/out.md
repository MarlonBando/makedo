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
<!-- outr ^exact$ -->

```bash
echo "Hello world!"
```
<!-- outr ^Hello world!$ -->

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
<!-- outr test\.log -->

```bash
echo 'cost: $10'
```
<!-- outr \$10 -->

## Character Classes

Use `[...]` to match any character in the set.

```bash
echo "file123.txt"
```
<!-- outr file[0-9]+\.txt -->

```bash
echo "hello"
```
<!-- outr h[aeiou]llo -->

## Quantifiers

Common quantifiers: `*` (zero or more), `+` (one or more), `?` (zero or one).

```bash
echo "goooal"
```
<!-- outr go+al -->

```bash
echo "color"
```
<!-- outr colou?r -->

## Alternation

Use `|` for OR patterns.

```bash
echo "cat"
```
<!-- outr (cat|dog) -->

## Real-World Example: Version Checking

Check if git version is 2.x or higher.

```bash
git --version
```
<!-- outr ^git version [2-9]\. -->

## Multiline Output

By default, newlines are trimmed from the end. Use `(?s)` flag or match newlines explicitly.

```bash
printf "line1\nline2\nline3"
```
<!-- outr line1\nline2\nline3 -->

## Empty Output

Match empty output with `^$`.

```bash
true
```
<!-- outr ^$ -->

## Whitespace Matching

Use `\s` for whitespace, `\t` for tabs.

```bash
echo "hello   world"
```
<!-- outr hello\s+world -->

