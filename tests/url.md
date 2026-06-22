# URL Checking

Makedo can verify that http(s) links resolve when invoked with `--check-urls`.
This test launches a local server in the background, generates a fixture
markdown with one working and one broken link, runs `makedo test --check-urls`
on it, and asserts both the failure status and the failure message.

## Start a local server

```bash
python3 -m http.server 8867 2>&1
```
`<!-- cmd curl -s http://localhost:8867 > /dev/null -->`

## Generate fixture, run makedo on it, clean up

```bash
cat > /tmp/url-fixture.md <<'EOF'
# Fixture

- [good](http://localhost:8867/)
- [broken](http://localhost:8867/does-not-exist)
EOF

makedo test --check-urls /tmp/url-fixture.md; echo "exit=$?"
rm /tmp/url-fixture.md
```
`<!-- out HTTP 404 -->`
`<!-- out exit=1 -->`
`<!-- out 2 URLs checked -->`
