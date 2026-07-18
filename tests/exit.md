# Exit Directive Tests

## Basic exit code assertion
Should pass when exit code matches.

```bash
bash -c 'exit 1'
```
<!-- exit 1 -->

## Basic exit code 0 assertion
Should pass when explicitly checking for 0.

```bash
bash -c 'exit 0'
```
<!-- exit 0 -->

## Negated exit assertion
Should pass when exit code does not match.

```bash
bash -c 'exit 1'
```
<!-- !exit 0 -->

## Negated exit assertion with different exit code
Should pass when exit code does not match.

```bash
bash -c 'exit 2'
```
<!-- !exit 1 -->

## Multiple directives with exit
Should pass when all match.

```bash
echo "hello"
bash -c 'exit 3'
```
<!-- out hello -->
<!-- exit 3 -->

## Negated exit with output directive
Should pass when all match.

```bash
echo "world"
bash -c 'exit 5'
```
<!-- out world -->
<!-- !exit 0 -->

## Expected failure with specific error output
Should pass when the command fails and outputs the expected error message.

```bash
bash -c 'echo "critical error"; exit 1'
```
<!-- !exit 0 -->
<!-- out critical error -->
