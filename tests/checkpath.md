# checkpath directive tests

Test the checkpath directive for asserting file and directory existence.

## Test 1: Check existing file

```bash
touch test_file_exists.txt
```
<!-- checkpath test_file_exists.txt -->
<!-- cmd rm test_file_exists.txt -->

## Test 2: Check existing directory

```bash
mkdir -p test_dir_exists
```
<!-- checkpath test_dir_exists -->
<!-- cmd rm -rf test_dir_exists -->

## Test 3: Negation of non-existing file

```bash
echo "Testing negation"
```
<!-- !checkpath definitely_does_not_exist_12345.txt -->

## Test 4: Verify failing checkpath (should fail if not negated and file doesn't exist)

```bash
echo "This test is negated structurally to test the failure case"
```
<!-- !checkpath non_existent_file.txt -->
