# Pseudo-Terminal (PTY) Edge Cases

This test suite ensures that our shared persistent shell handles PTY initialization edge cases correctly, specifically cross-platform terminal `ECHO` disabled states and interactive prompt suppression.

## 1. Terminal ECHO is Disabled
If `ECHO` is left ON (either because `stty -echo` failed or bash `readline` overwrote it), the PTY will echo the command `echo "exact"` back into the output stream before executing it. This test verifies that strictly matching `^exact$` succeeds, proving the command itself was not echoed.

```bash
echo "exact"
```
<!-- outr ^exact$ -->

## 2. Interactive Prompts are Suppressed
Because we attach a PTY, bash thinks it is interactive. It would normally print `PS1` and `PS2` prompts. A multi-line heredoc triggers the `PS2` prompt (usually `> `). This test verifies that `PS2` is fully disabled and doesn't leak into the output buffer.

```bash
cat << 'EOF' > /tmp/makedo_pty_test.txt
line 1
line 2
EOF
```
<!-- outr ^$ -->

## 3. Carriage Returns are Stripped
PTYs automatically translate `\n` to `\r\n`. We must ensure that our `CleanOutput` utility perfectly strips out `\r` so that our regex matching doesn't fail on hidden characters.

```bash
printf "line1\nline2"
```
<!-- outr ^line1\nline2$ -->

## 4. ANSI Escape Codes are Stripped
Commands run inside a PTY often detect the terminal and output rich ANSI colors. We ensure that `CleanOutput` correctly strips out ANSI codes before applying directive checks.

```bash
printf "\x1b[32mcolorful\x1b[0m"
```
<!-- outr ^colorful$ -->
