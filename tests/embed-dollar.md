# The `embed` Command for Dollar Format

These tests verify the `makedo embed` command handles the new dollar format (inline output updates) correctly. We use temporary files here to avoid modifying the test file itself.

## Embedding New Output

Verify that `embed` correctly injects new inline output into a dollar format block that currently has none.

````bash
cat << 'INNER_EOF' > /tmp/test_embed_dollar_1.md
```bash
$ echo "hello embed"
```
INNER_EOF

makedo embed /tmp/test_embed_dollar_1.md > /dev/null
````
<!-- cmd grep -q '\$ echo "hello embed"' /tmp/test_embed_dollar_1.md -->
<!-- cmd grep -q '^hello embed$' /tmp/test_embed_dollar_1.md -->
<!-- cmd ! grep -q '```stdout' /tmp/test_embed_dollar_1.md -->

## Updating Existing Output

Verify that `embed` replaces old inline output with new inline output.

````bash
cat << 'INNER_EOF' > /tmp/test_embed_dollar_2.md
```bash
$ echo "new data"
old data
```
INNER_EOF

makedo embed /tmp/test_embed_dollar_2.md > /dev/null
````
<!-- cmd grep -q '^new data$' /tmp/test_embed_dollar_2.md -->
<!-- cmd ! grep -q '^old data$' /tmp/test_embed_dollar_2.md -->

## Handling Failures

Verify that `embed` leaves the old inline output intact if the command fails.

````bash
cat << 'INNER_EOF' > /tmp/test_embed_dollar_3.md
```bash
$ exit 1
keep me
```
INNER_EOF

makedo embed /tmp/test_embed_dollar_3.md > /dev/null
````
<!-- cmd grep -q '^keep me$' /tmp/test_embed_dollar_3.md -->

## Cleanup

````bash
rm -f /tmp/test_embed_dollar_1.md /tmp/test_embed_dollar_2.md /tmp/test_embed_dollar_3.md
````
<!-- cmd test ! -f /tmp/test_embed_dollar_1.md -->
