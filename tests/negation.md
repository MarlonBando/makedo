# Negation Tests

This file tests negative assertions using the `!out`, `!cmd`, and `!pwd` directives.

## Not Out

The output is "hello", so `!out world` should pass.

```bash
echo hello
```
<!-- !out world -->
<!-- out hello -->

It should also handle regex negations.

```bash
echo 123
```
<!-- !outr [a-z]+ -->
<!-- outr [0-9]+ -->

## Not Cmd

The command exits with an error (1), so `!cmd` should pass.

```bash
echo testing not cmd
```
<!-- !cmd false -->
<!-- cmd true -->

## Not Pwd

The current directory is NOT `/etc`, so this should pass.

```bash
echo testing not pwd
```
<!-- !pwd /etc -->

## Failures and Type substitutions

Even with type substitution, it should correctly negate. We check that an output is NOT the UUID format.

```bash
echo "not a uuid"
```
<!-- !outr ${{uuid}} -->

All tests should pass.