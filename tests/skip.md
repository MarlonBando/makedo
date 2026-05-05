# Skip Directive Tests

These tests verify that `<!-- skip -->` keeps shell blocks as regular markdown and prevents makedo processing.

## Skip Prevents Execution

````bash
cat << 'EOF' > /tmp/makedo_skip_case_1.md
```bash
echo "SHOULD_NOT_RUN" > /tmp/makedo_skip_case_1.out
```
<!-- skip -->
EOF

makedo test /tmp/makedo_skip_case_1.md
````
<!-- out 0/0 tests passed -->
<!-- cmd test ! -f /tmp/makedo_skip_case_1.out -->

## Skip Wins Over Other Directives

````bash
cat << 'EOF' > /tmp/makedo_skip_case_2.md
```bash
echo "SHOULD_NOT_RUN" > /tmp/makedo_skip_case_2.out
```
<!-- out SHOULD_NOT_RUN -->
<!-- cmd test -f /tmp/makedo_skip_case_2.out -->
<!-- skip -->
EOF

makedo test /tmp/makedo_skip_case_2.md
````
<!-- out 0/0 tests passed -->
<!-- cmd test ! -f /tmp/makedo_skip_case_2.out -->

## `!skip` Is Invalid (Does Not Skip)

````bash
cat << 'EOF' > /tmp/makedo_skip_case_3.md
```bash
echo "SETUP_RAN" > /tmp/makedo_skip_case_3.out
```
<!-- !skip -->
EOF

makedo test /tmp/makedo_skip_case_3.md
````
<!-- out 0/0 tests passed -->
<!-- cmd test -f /tmp/makedo_skip_case_3.out -->

## Cleanup

````bash
rm -f /tmp/makedo_skip_case_1.md /tmp/makedo_skip_case_2.md /tmp/makedo_skip_case_3.md
rm -f /tmp/makedo_skip_case_1.out /tmp/makedo_skip_case_2.out /tmp/makedo_skip_case_3.out
````
<!-- cmd test ! -f /tmp/makedo_skip_case_1.md -->
<!-- cmd test ! -f /tmp/makedo_skip_case_2.md -->
<!-- cmd test ! -f /tmp/makedo_skip_case_3.md -->
