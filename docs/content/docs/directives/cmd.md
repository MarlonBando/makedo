---
title: "cmd"
weight: 2
bookToc: true
---

# The `cmd` Directive

The `cmd` directive runs a verification command after the code block executes. The test passes if the verification command exits with code 0.

## Basic File Checking

Verify a file exists after creation:

```bash
touch /tmp/makedo_doc_cmd_test.txt
```
<!-- cmd test -f /tmp/makedo_doc_cmd_test.txt -->

Verify a file was deleted:

```bash
rm -f /tmp/makedo_doc_cmd_test.txt
```
<!-- cmd test ! -f /tmp/makedo_doc_cmd_test.txt -->

## Directory Checking

```bash
mkdir -p /tmp/makedo_doc_cmd_dir
```
<!-- cmd test -d /tmp/makedo_doc_cmd_dir -->

```bash
rmdir /tmp/makedo_doc_cmd_dir
```
<!-- cmd test ! -d /tmp/makedo_doc_cmd_dir -->

## Tool Availability

Check if a command is installed:

```bash
echo "Checking for git..."
```
<!-- cmd git --version -->

```bash
echo "Checking for bash..."
```
<!-- cmd which bash -->

## File Content Verification

Use `grep` to verify file contents:

```bash
echo "success" > /tmp/makedo_doc_content.txt
```
<!-- cmd grep -q "success" /tmp/makedo_doc_content.txt -->

Verify a file does NOT contain specific text:

```bash
echo "pass" > /tmp/makedo_doc_content2.txt
```
<!-- cmd ! grep -q "fail" /tmp/makedo_doc_content2.txt -->

## Cleanup

```bash
rm -f /tmp/makedo_doc_content.txt /tmp/makedo_doc_content2.txt
```
<!-- cmd test ! -f /tmp/makedo_doc_content.txt -->

## Negation

Use `!cmd` to assert a command fails:

```bash
echo "testing negated cmd"
```
<!-- !cmd false -->
<!-- cmd true -->
