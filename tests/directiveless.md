# Directiveless Blocks Tests

This test verifies that directiveless blocks implicitly check the exit code, count as tests, and do not abort the execution walk upon failure.

## 1. Mixed Blocks Test

We create a markdown file containing three blocks:
1. A directiveless block that fails (`exit 1`).
2. A directiveless block that succeeds (`exit 0`).
3. A block with directives that succeeds.

`makedo test` should evaluate all three and report 2 passed, 1 failed.

````bash
cat << 'EOF' > /tmp/test_directiveless.md
```bash
false
```

```bash
echo "success"
```

```bash
echo "with directive"
```
<!-- out directive -->
EOF

# Run makedo test. It will return non-zero exit code because of the failure.
# We capture the output and ensure it doesn't abort early.
bin/makedo test /tmp/test_directiveless.md > /tmp/test_directiveless.out 2>&1 || true
````
<!-- cmd grep -q '2/3 tests passed — 1 block(s) failed' /tmp/test_directiveless.out -->

## Cleanup

````bash
rm -f /tmp/test_directiveless.md /tmp/test_directiveless.out
````
<!-- !checkpath /tmp/test_directiveless.md -->
