# Multiple Directives

A single code block can have multiple directives. The block executes once, then each directive verifies the result in order.

## Combining `out` and `cmd`

Verify both output and side effects from the same command.

```bash
mkdir -p /tmp/makedo_multi_1
echo "Directory created"
```
<!-- out created -->
<!-- cmd test -d /tmp/makedo_multi_1 -->

## Multiple `cmd` Directives

Run several verification commands on the same block.

```bash
touch /tmp/makedo_multi_file.txt
chmod +x /tmp/makedo_multi_file.txt
```
<!-- cmd test -f /tmp/makedo_multi_file.txt -->
<!-- cmd test -x /tmp/makedo_multi_file.txt -->

## Order Matters

Directives are processed in the order they appear.

```bash
echo "test output"
```
<!-- out test -->
<!-- out output -->
<!-- cmd test 1 -eq 1 -->

## Verify Output and File Content

Create a file and verify both stdout and file contents.

```bash
echo "success" | tee /tmp/makedo_tee.txt
```
<!-- out success -->
<!-- cmd test -f /tmp/makedo_tee.txt -->
<!-- cmd grep -q "success" /tmp/makedo_tee.txt -->

## Complex Example: File Creation Pipeline

Create multiple files and verify each step.

```bash
mkdir -p /tmp/makedo_pipeline
echo "data" > /tmp/makedo_pipeline/input.txt
cat /tmp/makedo_pipeline/input.txt > /tmp/makedo_pipeline/output.txt
echo "Pipeline complete"
```
<!-- out complete -->
<!-- cmd test -d /tmp/makedo_pipeline -->
<!-- cmd test -f /tmp/makedo_pipeline/input.txt -->
<!-- cmd test -f /tmp/makedo_pipeline/output.txt -->
<!-- cmd grep -q "data" /tmp/makedo_pipeline/output.txt -->

## Non-Directive Comments Stop Collection

Only consecutive directive comments are collected. Regular HTML comments stop the collection.

```bash
echo "hello world"
```
<!-- out hello -->
<!-- cmd test 1 -eq 1 -->
<!-- This is just a regular comment -->
<!-- cmd this will NOT be processed -->

## Verify Tool Installation and Version

Check that a tool exists and outputs the correct version pattern.

```bash
git --version
```
<!-- out git version ${{version}} -->
<!-- cmd git --version -->

## File Permissions Example

Create file, set permissions, verify both existence and permissions.

```bash
echo "content" > /tmp/makedo_perms.txt
chmod 644 /tmp/makedo_perms.txt
```
<!-- cmd test -f /tmp/makedo_perms.txt -->
<!-- cmd test -r /tmp/makedo_perms.txt -->
<!-- cmd test -w /tmp/makedo_perms.txt -->
<!-- cmd test ! -x /tmp/makedo_perms.txt -->

## Cleanup: Remove All Test Files

Clean up all test files and directories created in this documentation.

```bash
rm -rf /tmp/makedo_multi_1
rm -f /tmp/makedo_multi_file.txt
rm -f /tmp/makedo_tee.txt
rm -rf /tmp/makedo_pipeline
rm -f /tmp/makedo_perms.txt
```
<!-- cmd test ! -d /tmp/makedo_multi_1 -->
<!-- cmd test ! -f /tmp/makedo_multi_file.txt -->
<!-- cmd test ! -f /tmp/makedo_tee.txt -->
<!-- cmd test ! -d /tmp/makedo_pipeline -->
<!-- cmd test ! -f /tmp/makedo_perms.txt -->

## Best Practices

1. **Execute once, verify many**: The code block runs once, keeping tests fast
2. **Order directives logically**: Put `out` first, then `cmd` checks
3. **Clean up resources**: Always remove temporary files/directories
4. **Stop at non-directives**: Regular comments break the directive chain
5. **Use specific checks**: Multiple simple `cmd` directives are better than one complex one
