# Background Commands

Makedo can run long-running commands in the background. When a command with directives passes its checks, it stays alive for subsequent blocks to use.

## Start a Server

Use a `cmd` directive to probe when the server is ready:

```bash
python3 -m http.server 8866 2>&1 &
```
<!-- cmd curl -s http://localhost:8866 > /dev/null 2>&1 -->

## Interact with the Server

The server is still running. We can make requests:

```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:8866
```
<!-- out 200 -->

At the end of the document, makedo automatically kills the server process.

```bash
echo "Hello world" &
```
<!-- out Hello world -->
