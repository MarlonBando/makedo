---
title: "Background Processes"
weight: 5
bookToc: true
---

# Background Processes

MakeDo automatically manages background processes. When a code block with directives passes its checks before the command exits, MakeDo keeps the command running in the background for subsequent blocks to use.

## How It Works

1. **Start**: The shell command is started as a new process group
2. **Continuous Evaluation**: As output arrives, MakeDo evaluates directives continuously
3. **Ready State**: If all directives pass before the command exits, MakeDo moves on to the next block while leaving the command running
4. **Automatic Cleanup**: At the end of the document, MakeDo kills all background processes

## Example: Starting a Server

You can start a long-running server and immediately use it in the next block:

```bash
python3 -m http.server 8877 2>&1
```
<!-- cmd curl -s http://localhost:8877 > /dev/null 2>&1 -->

The server is now running. You can interact with it:

```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:8877
```
<!-- out 200 -->

At the end of this page, MakeDo automatically kills the server.

## Stall Detection

If a command produces no output for 10 seconds (the default `StallTimeout`), it is marked as **Stalled**. If the block had directives that haven't passed yet, this results in a test failure.

## Background Operator Warning

MakeDo manages background processes automatically based on directive satisfaction. Using the shell `&` operator manually is **discouraged**. If MakeDo detects a lone `&` operator, it will issue a warning.
