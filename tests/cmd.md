# The `cmd` Directive

The `cmd` directive runs a verification command after the code block executes. The test passes if the verification command exits with code 0.
Add after a code block a html comment in this format `<!-- cmd ${your_command} -->`

## Basic File Checking

Verify a file exists after creation.

```bash
touch /tmp/makedo_test_file.txt
```
<!-- cmd test -f /tmp/makedo_test_file.txt -->

Verify a file was deleted.

```bash
rm -f /tmp/makedo_test_file.txt
```
<!-- cmd test ! -f /tmp/makedo_test_file.txt -->

## Directory Checking

Create and verify a directory exists.

```bash
mkdir -p /tmp/makedo_test_dir
```
<!-- cmd test -d /tmp/makedo_test_dir -->

Verify directory removal.

```bash
rmdir /tmp/makedo_test_dir
```
<!-- cmd test ! -d /tmp/makedo_test_dir -->

## File Properties

Check if a file is executable.

```bash
echo '#!/bin/bash' > /tmp/makedo_script.sh
chmod +x /tmp/makedo_script.sh
```
<!-- cmd test -x /tmp/makedo_script.sh -->

Check if a file is readable.

```bash
echo "content" > /tmp/makedo_readable.txt
```
<!-- cmd test -r /tmp/makedo_readable.txt -->

Clean up test files.

```bash
rm -f /tmp/makedo_script.sh /tmp/makedo_readable.txt
```
<!-- cmd test ! -f /tmp/makedo_script.sh -->

## Tool Availability

Check if git is installed.

```bash
echo "Checking for git..."
```
<!-- cmd git --version -->

Check if a command exists using `which`.

```bash
echo "Checking for bash..."
```
<!-- cmd which bash -->

## Network Availability

Check if localhost is reachable (requires netcat/nc).

```bash
echo "Testing network tools..."
```
<!-- cmd ping -c 1 127.0.0.1 -->

## File Content Verification

Verify file contains specific text using grep.

```bash
echo "success" > /tmp/makedo_content.txt
```
<!-- cmd grep -q "success" /tmp/makedo_content.txt -->

Verify file does NOT contain text.

```bash
echo "pass" > /tmp/makedo_content2.txt
```
<!-- cmd ! grep -q "fail" /tmp/makedo_content2.txt -->

Clean up content test files.

```bash
rm -f /tmp/makedo_content.txt /tmp/makedo_content2.txt
```
<!-- cmd test ! -f /tmp/makedo_content.txt -->

## Custom Exit Codes

Any command that exits with 0 on success works.

```bash
echo "Custom check"
```
<!-- cmd bash -c "exit 0" -->

