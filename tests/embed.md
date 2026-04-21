# The `embed` Command

These tests verify the `makedo embed` command, which executes code blocks and injects their output back into the file. We use temporary files here to avoid modifying the test file itself.

## Embedding New Output

Verify that `embed` adds a new stdout block when none exists.

````bash
cat << 'EOF' > /tmp/test_embed_1.md
```bash
echo "hello embed"
```
<!-- out hello embed -->
EOF

makedo embed /tmp/test_embed_1.md > /dev/null
````
<!-- cmd grep -q '```stdout' /tmp/test_embed_1.md -->
<!-- cmd grep -q 'hello embed' /tmp/test_embed_1.md -->

## Updating Existing Output

Verify that `embed` replaces old output with new output.

````bash
cat << 'EOF' > /tmp/test_embed_2.md
```bash
echo "new data"
```
<!-- out new data -->

```stdout
old data
```
EOF

makedo embed /tmp/test_embed_2.md > /dev/null
````
<!-- cmd grep -q 'new data' /tmp/test_embed_2.md -->
<!-- cmd ! grep -q 'old data' /tmp/test_embed_2.md -->

## Handling Failures

Verify that `embed` leaves the old output intact if the command fails.

````bash
cat << 'EOF' > /tmp/test_embed_3.md
```bash
exit 1
```
<!-- out never_match -->

```stdout
keep me
```
EOF

makedo embed /tmp/test_embed_3.md > /dev/null
````
<!-- cmd grep -q 'keep me' /tmp/test_embed_3.md -->

## Cleanup

````bash
rm -f /tmp/test_embed_1.md /tmp/test_embed_2.md /tmp/test_embed_3.md
````
<!-- cmd test ! -f /tmp/test_embed_1.md -->
