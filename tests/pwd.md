# pwd directive tests

Test the pwd directive with simplified wildcard patterns.
Only supports a single asterisk (*) at the beginning or end of the pattern.

## Test 1: Exact suffix match (no wildcard)

```bash
echo "Path must end with 'makedo'"
```
<!-- pwd makedo -->

## Test 2: Ends with pattern

```bash
echo "Path must end with 'proj/makedo'"
```
<!-- pwd proj/makedo -->

## Test 3: Starts with pattern

```bash
echo "Path must start with '/home'"
```
<!-- pwd /home* -->

## Test 4: Ends with pattern (suffix check)

```bash
echo "Path must end with 'thesis/proj/makedo'"
```
<!-- pwd *thesis/proj/makedo -->

## Failing test cases (invalid patterns)

### Test 5: Multiple wildcards (should fail with descriptive error)

```bash
echo "This pattern has multiple wildcards and should fail"
```
<!-- pwd */proj*/makedo -->

### Test 6: Wildcard in the middle (should fail with descriptive error)

```bash
echo "This pattern has wildcard in middle and should fail"
```
<!-- pwd /home/*thesis/makedo -->

### Test 7: Contains pattern (should fail with descriptive error)

```bash
echo "This pattern tries to match 'contains' and should fail"
```
<!-- pwd *proj* -->

