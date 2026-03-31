# No Orphan Process Test

Verify that makedo properly cleans up background processes after test execution.

## Precondition: Port Must Be Free

Ensure port 8866 is not already in use before running the test:

```bash
! lsof -i :8866 -t 2>/dev/null
```
<!-- out  -->

## Run Background Test

Execute the background test which starts a Python HTTP server on port 8866:

```bash
./bin/makedo test tests/background.md
```
<!-- out 2/2 tests passed -->

## Verify No Orphan Process

After makedo finishes, the port should be free again:

```bash
! lsof -i :8866 -t 2>/dev/null
```
<!-- out  -->
