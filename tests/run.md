# Test `run` Command Output

This test verifies that the `makedo run` command correctly streams user code output to the terminal without leaking internal execution markers (like `MAKEDO_DONE`) or state dumps (like `pwd` and `export`).

## Setup

First, we create a temporary markdown file containing a simple code block. Notice we escape the inner code fence so it doesn't break this file.

```bash
cat << 'EOF' > /tmp/test_run.md
```bash
echo "Hello from streaming"
\```
EOF
```
<!-- outr ^$ -->

## Execute and Verify

We execute the `run` command on the temporary file and inspect its standard output stream.

```bash
bin/makedo run /tmp/test_run.md
```
<!-- out Hello from streaming -->
<!-- !out MAKEDO_DONE -->
<!-- !out export -p -->
<!-- !out pwd > -->

## Cleanup

Clean up the temporary markdown file.

```bash
rm -f /tmp/test_run.md
```
<!-- outr ^$ -->
