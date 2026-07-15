# Background Parser Edge Cases

This test demonstrates the current limitations of `hasBackgroundOperator` which uses basic string parsing instead of an AST.
Because these edge cases falsely trigger the background detector, the engine runs them in `executeBackground` instead of the persistent shell, causing any exported variables to be lost!

## Test 1: Stderr Redirection with space (`>& 2`)

```bash
export EDGE_VAR_1="hello"
echo "redirection" >& 2
```
<!-- out redirection -->

Because this falsely triggered the background branch, `EDGE_VAR_1` will be lost in the persistent shell!

```bash
echo "VAR1: $EDGE_VAR_1"
```
<!-- out VAR1: hello -->


## Test 2: Subshell with background process

```bash
export EDGE_VAR_2="world"
val=$(echo "subshell" & wait)
echo $val
```
<!-- out subshell -->

Again, this falsely triggered the background branch, so `EDGE_VAR_2` will be lost!

```bash
echo "VAR2: $EDGE_VAR_2"
```
<!-- out VAR2: world -->

## Test 3: Ampersand in Quotes (Supported)

```bash
export EDGE_VAR_3="quotes"
echo "this string has an & inside it"
```
<!-- out this string has an & inside it -->

Because our parser properly ignores strings, this stays in the persistent shell:

```bash
echo "VAR3: $EDGE_VAR_3"
```
<!-- out VAR3: quotes -->


## Test 4: Ampersand in Comments (Supported)

```bash
export EDGE_VAR_4="comments"
# this comment has an & inside it
echo "comment test"
```
<!-- out comment test -->

Because our parser strips comments, this stays in the persistent shell:

```bash
echo "VAR4: $EDGE_VAR_4"
```
<!-- out VAR4: comments -->


## Test 5: Standard Redirection `2>&1` (Supported)

```bash
export EDGE_VAR_5="redirection"
echo "redirect test" > /dev/null 2>&1
echo "done"
```
<!-- out done -->

Because `2>&1` does not have a space after the `&`, our parser correctly ignores it:

```bash
echo "VAR5: $EDGE_VAR_5"
```
<!-- out VAR5: redirection -->
