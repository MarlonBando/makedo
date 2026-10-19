# MakeDo: docs that can't lie, tests anyone can read

MakeDo runs the shell blocks in your Markdown and checks what they do. Add a [directive](https://marlonbando.github.io/makedo/docs/directives/)  under a code block, run `makedo test`, and the block becomes a test. Your docs render exactly as before.

**[Read the documentation →](https://marlonbando.github.io/makedo/)**

## Test your documentation

Add one comment under an example:

```bash
makedo --version
```
`<!-- out ${{version}} -->`

If a release renames the flag, `makedo test README.md` fails and CI goes red:

```text
FAIL README.md:12
│ makedo --version
│ ✗ <!-- out ${{version}} -->
│     expected: ${{version}}
│     actual:   Error: unknown flag: --version

✗ 7/8 tests passed — 1 block(s) failed
```

`${{version}}` matches any version string, so the test doesn't break on every release.

## Test your CLI in Markdown

A test file is just a document. All blocks run in one shell session, so `cd`, variables and files carry over. Processes started with `&` keep running in the background and are stopped when the document ends. Besides output, you can check files, exit codes and the working directory.

```bash
cd "$(mktemp -d)"
mkdir site && echo '<title>Home</title>' > site/index.html
```
`<!-- checkpath site/index.html -->`

```bash
python3 -m http.server 8765 --directory site &
```
`<!-- out Serving HTTP -->`

```bash
curl -s localhost:8765
```
`<!-- out <title>Home</title> -->`

MakeDo waits until the server prints `Serving HTTP`, then moves on to the next block while the server keeps running.

## Keep your agent instructions working

AI coding agents follow `AGENTS.md` and skill files to the letter. If those files tell an agent to run a command that no longer exists, the agent fails too. Run them through MakeDo:

````markdown
Before committing, run the linter:

```bash
make lint
```
````

A block without directives must still exit with 0, so a removed `make` target is caught:

```text
FAIL AGENTS.md:4
│ make lint
│ exit 2
│ --- output (tail) ---
│ make: *** No rule to make target 'lint'.  Stop.
```

## Evidence

MakeDo was built and evaluated as an MSc thesis at the Technical University of Denmark (DTU). The main findings:

- **Docs break and stay broken.** Over 9 years, Hugo's quick-start guide broke 6 times and stayed broken for 798 days in total. One breakage lasted 520 days. In a survey of 62 popular repositories, 79% have shell examples in their docs and 67% of those never test them.
- **MakeDo finds real bugs.** We changed 19 code fences in ripgrep's `GUIDE.md` from ` ``` ` to ` ```bash ` and ran `makedo test` once. It found 6 broken examples nobody had reported ([#3448](https://github.com/BurntSushi/ripgrep/issues/3448)). Replaying the commit that broke them showed a MakeDo CI check would have failed before the merge.
- **It works as a test framework.** We ported 70 of the 75 tests in Hugo's integration suite to MakeDo. The port caught the same 8 of 10 injected bugs as Hugo's original `testscript` suite. It needed 12 lines of Bash helpers instead of 352 lines of Go.

**[Read the thesis →](https://marlonbando.github.io/makedo/docs/thesis/)**

## Install

Download the zip for your platform from the [latest release](https://github.com/MarlonBando/makedo/releases/latest). Each zip contains a single `makedo` binary: unzip it and move it somewhere on your `PATH`.

```bash
VERSION=0.1.0
OS=linux      # or darwin
ARCH=amd64    # or arm64
curl -LO "https://github.com/MarlonBando/makedo/releases/download/${VERSION}/makedo_${VERSION}_${OS}_${ARCH}.zip"
unzip "makedo_${VERSION}_${OS}_${ARCH}.zip"
sudo mv makedo /usr/local/bin/    # or: mv makedo ~/.local/bin/
makedo --version
```
<!-- skip -->

Or build from source (requires Go 1.24+):

```bash
git clone https://github.com/MarlonBando/makedo.git
cd makedo
make build
```

> **Note:** MakeDo supports bash only. On Windows, use [WSL](https://learn.microsoft.com/windows/wsl/install): install MakeDo inside your WSL distribution with the steps above, then run it from PowerShell in your project folder:
>
> ```powershell
> wsl makedo test README.md
> ```

## Quick Start

Create `hello.md`:

````bash
cat > hello.md <<'EOF'
```bash
echo "hello world"
```
`<!-- out hello world -->`
EOF
````
`<!-- checkpath hello.md -->`

Run it:

```bash
makedo test hello.md
```
`<!-- out 1/1 tests passed -->`

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

## Limitations

- **One output stream.** Commands run in a pseudo-terminal, as they would for a reader, so stdout and stderr are merged. You can't assert which stream a message went to.
- **No platform-specific blocks.** A block can't be set to run only on one operating system.
- **Bash only.** On Windows, use WSL.

## Tested with MakeDo

MakeDo tests itself. Its test suite in [`tests/`](tests/) is 21 Markdown files, every page of the [documentation](https://marlonbando.github.io/makedo/) asserts its own examples, and this README runs in CI too.
