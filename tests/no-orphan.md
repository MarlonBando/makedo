# No Orphan Process Test

Verify that makedo properly cleans up background processes after test execution.

## Precondition

Ensure port 8866 is free and no residual sleeps exist:

```bash
! lsof -i :8866 -t 2>/dev/null
pkill -f "sleep 9999" || true
```
<!-- out  -->

## Setup and Run 

We'll run a test that spawns both a long-lived server and a background sleep task (which acts as an infinite short/long script).

```bash
cat << 'INNER_EOF' > test-orphan.md
```bash
python3 -m http.server 8866 2>&1 &
```
<!-- cmd curl -s http://localhost:8866 > /dev/null 2>&1 -->

```bash
sleep 9999 &
```
INNER_EOF

bin/makedo test test-orphan.md
```
<!-- out 2/2 tests passed -->

## Verify No Orphan Processes

After `makedo` finishes, the inner registry should have SIGKILLed both the server and the sleep command. 

```bash
# Give the OS time to reap the SIGKILLed processes (retry loop)
for i in {1..10}; do
  if ! pgrep -f "[s]leep 9999" > /dev/null && ! lsof -i :8866 -t 2>/dev/null > /dev/null; then
    exit 0
  fi
  sleep 0.5
done
echo "Processes still alive!"
exit 1
```
<!-- out  -->

## Cleanup

```bash
rm test-orphan.md
```
