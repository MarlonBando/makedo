---
title: "Getting Started"
weight: 1
bookToc: true
---

# Getting Started

MakeDo turns your Markdown documentation into executable test suites. Every fenced code block with a shell language identifier (`bash`, `sh`, `zsh`, or `shell`) is automatically detected and can be tested.

## Installation

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
sudo cp bin/makedo /usr/local/bin/
```
<!-- skip -->

> **Note:** MakeDo supports bash only. On Windows, use [WSL](https://learn.microsoft.com/windows/wsl/install): install MakeDo inside your WSL distribution with the steps above, then run it from PowerShell in your project folder:
>
> ```powershell
> wsl makedo test README.md
> ```

Check that it's installed by printing the version:

```bash
makedo -v
```
`<!-- out ${{version}} -->`

```stdout
${{version}}
```

## Available Commands

MakeDo provides three main commands:

| Command | Description |
| :--- | :--- |
| `test` | Run code blocks and verify directives pass |
| `embed` | Run code blocks and inject output into the Markdown |
| `run` | Simply execute all code blocks |

You can see all available commands with:

```bash
makedo -h
```
`<!-- out Available Commands -->`
`<!-- out test -->`
`<!-- out embed -->`
`<!-- out run -->`

```stdout
A markdown-based task runner

Usage:
  makedo [flags]
  makedo [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  embed       Run fenced code blocks and embed output into the markdown file
  help        Help about any command
  run         Run fenced code blocks from a markdown file
  test        Test code blocks in your markdown file

Flags:
  -h, --help      help for makedo
  -v, --version   Print version

Use "makedo [command] --help" for more information about a command.
```

## Your First Test

Create a Markdown file called `hello.md` with a code block and a directive.

You can do it by copying the content of this block into the file:
````markdown
```bash
echo hello world!
```
<!-- out hello world! -->
````

or by running this command in your terminal:
```bash
printf '```bash\necho hello world!\n```\n<!-- out hello world! -->\n' > hello.md
```
`<!-- checkpath hello.md -->`

Then run:

```bash
makedo test hello.md
```
`<!-- out 1/1 -->`

```stdout
hello.md:2 ✓

✓ 1/1 tests passed
```

MakeDo will execute the code block and verify the output matches the directive.

## A Practical Example

For a real-world example you don't have to look far: this documentation is itself tested with MakeDo. Every page contains directives that assert the behaviour it describes. Download the raw Markdown of the Dynamic Types page:

```bash
curl -sSLO https://raw.githubusercontent.com/MarlonBando/makedo/refs/heads/main/docs/content/docs/dynamic-types.md
```
`<!-- cmd grep -q '# Dynamic Types' dynamic-types.md -->`

Open the file in your favourite editor and look at how every behaviour is asserted.

Now run MakeDo on it:

```bash
makedo test dynamic-types.md
```
`<!-- exit 0 -->`
`<!-- out tests passed -->`

```stdout
dynamic-types.md:35 ✓
dynamic-types.md:42 ✓
dynamic-types.md:50 ✓
dynamic-types.md:66 ✓

✓ 4/4 tests passed
```

## Run It in CI

Add MakeDo to your GitHub Actions workflow to test your docs on every push:

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

## Cleanup

Remove the files created in this guide:

```bash
rm -f hello.md dynamic-types.md
```
`<!-- !checkpath hello.md -->`
`<!-- !checkpath dynamic-types.md -->`
