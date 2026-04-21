# pwd directive tests

Test the pwd directive with simplified wildcard patterns.
Only supports a single asterisk (*) at the beginning or end of the pattern.

## Test 1: Exact suffix match (no wildcard)

```bash
echo "Path must end with 'makedo'"
```
<!-- pwd makedo -->

## Test 3: Starts with pattern

```bash
echo "Path must start with '/home'"
```
<!-- pwd /home* -->
