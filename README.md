# MakeDo: test your CLI with Markdown

Write your CLI's tests in Markdown. The docs are the tests, and the tests are docs anyone can read.

**[Read the documentation →](https://marlonbando.github.io/makedo/)**

Add a directive below a shell code block and MakeDo checks it:

```bash
makedo -v
```
`<!-- out ${{version}} -->`

Run `makedo test README.md` and every example in this file is executed and verified.

## Why MakeDo

- **Docs that can't lie.** When an example stops working, the test fails.
- **CLI tests without a test framework.** No harness to write, just commands and the output you expect, in plain Markdown.
- **A real shell session.** Variables, `cd` and files carry over between blocks. Servers started with `&` keep running in the background and are cleaned up at the end.
- **Output that writes itself.** `makedo embed` runs your blocks and writes their real output into the doc.

## Install

Build from source (requires Go 1.24+):

```bash
git clone https://github.com/MarlonBando/makedo.git
cd makedo
make build
sudo cp bin/makedo /usr/local/bin/
```
<!-- skip -->

> **Note:** MakeDo supports Linux only. On Windows, use [WSL](https://learn.microsoft.com/windows/wsl/install): install MakeDo inside your WSL distribution with the steps above, then run it from PowerShell in your project folder:
>
> ```powershell
> wsl makedo test README.md
> ```

## Quick Start

Create `hello.md`:

````markdown
```bash
echo "hello world"
```
`<!-- out hello world -->`
````

Run it:

```bash
$ makedo test hello.md
hello.md:2 ✓

✓ 1/1 tests passed
```
<!-- skip -->
## Directives

Directives are HTML comments placed right after a code block. Wrap them in backticks to show them in the rendered page, or leave them bare to hide them.

| Directive | Passes when |
| :--- | :--- |
| `out` | the output contains a string |
| `outr` | the output matches a regular expression |
| `cmd` | a verification command exits with 0 |
| `exit` | the block exits with the given code |
| `pwd` | the current directory matches a pattern |
| `checkpath` | a file or directory exists |
| `skip` | never runs: the block is ignored |

Prefix any directive with `!` to negate it. For output that changes on every run, use placeholders like `${{date}}`, `${{uuid}}`, `${{ip}}` or `${{number}}`:

```bash
echo "Build finished on $(date +%F) in 1.42s"
```
`<!-- out Build finished on ${{date}} in ${{number}}s -->`

See the [directives reference](https://marlonbando.github.io/makedo/docs/directives/) and [dynamic types](https://marlonbando.github.io/makedo/docs/dynamic-types/) for details.

## Commands

| Command | What it does |
| :--- | :--- |
| `makedo test <file>` | Runs the code blocks and verifies their directives |
| `makedo embed <file>` | Runs the code blocks and writes their output into the file |
| `makedo run <file>` | Runs the code blocks, without checks |

## Run It in CI

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: '1.24.x'

- name: Install makedo
  run: |
    git clone --depth 1 https://github.com/MarlonBando/makedo.git /tmp/makedo
    cd /tmp/makedo/src
    go build -ldflags "-X main.versionString=$(cat ../VERSION)" -o "$HOME/.local/bin/makedo"
    echo "$HOME/.local/bin" >> "$GITHUB_PATH"

- name: Test docs
  run: for f in docs/*.md; do makedo test "$f"; done
```

## Tested with MakeDo

MakeDo tests itself. Its test suite in [`tests/`](tests/) is 21 Markdown files, every page of the [documentation](https://marlonbando.github.io/makedo/) asserts its own examples, and this README runs in CI too.
