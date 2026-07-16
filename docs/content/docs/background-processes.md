---
title: "Background Processes"
weight: 5
bookToc: true
---

# Background Processes

MakeDo supports running background processes. By using the shell `&` operator, you can start a long-running service. MakeDo will run the command in a detached background shell, monitor its output, and wait for its directives to pass before proceeding to the next block.

## How It Works

1. **Start**: The shell command is started as a new process group in the background
2. **Continuous Evaluation**: As output arrives, MakeDo evaluates directives continuously
3. **Ready State**: If all directives pass before the command exits, MakeDo moves on to the next block while leaving the command running
4. **Automatic Cleanup**: At the end of the document, MakeDo kills all background processes

## Example: Starting a Server

You can start a long-running server and immediately use it in the next block:

```bash
python3 -m http.server 8877 2>&1 &
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

## The Background Operator

You MUST use the standard shell background operator (`&`) to tell MakeDo that a command should run in the background. Without it, MakeDo will wait synchronously for the command to finish, which will cause your test to stall if it's a long-running server. MakeDo automatically manages the lifecycle of these background processes and ensures they are safely terminated at the end of the test.
