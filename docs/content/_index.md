---
title: "MakeDo"
type: docs
---

# MakeDo

**Docs that can't lie, tests anyone can read.**

MakeDo runs the shell blocks in your Markdown and checks what they do. Add a [directive]({{< relref "/docs/directives" >}}) under a code block, run `makedo test`, and the block becomes a test. Your docs render exactly as before.

See [Getting Started]({{< relref "/docs/getting-started" >}}) to install MakeDo and write your first test.

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

`${{version}}` matches any version string, so the test doesn't break on every release. See [Dynamic Types]({{< relref "/docs/dynamic-types" >}}) for the other placeholders.

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

MakeDo waits until the server prints `Serving HTTP`, then moves on to the next block while the server keeps running. See [Environment & State]({{< relref "/docs/environment" >}}) and [Background Processes]({{< relref "/docs/background-processes" >}}).

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

**[Read the thesis →]({{< relref "/docs/thesis" >}})**

## Limitations

- **One output stream.** Commands run in a pseudo-terminal, as they would for a reader, so stdout and stderr are merged. You can't assert which stream a message went to.
- **No platform-specific blocks.** A block can't be set to run only on one operating system.
- **Bash only.** On Windows, use WSL.

## Tested with MakeDo

MakeDo tests itself: its test suite is written in Markdown, and every page of this documentation asserts its own examples.
