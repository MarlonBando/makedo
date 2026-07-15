This test verifies that `makedo` detects the `&` operator and warns the user without interfering with normal test execution. We test each scenario individually.

### Test 1: Simple background op

````bash
cat << 'EOF' > test1.md
```bash
echo Hello World AMPERSAND_TOKEN
```
<!-- out Hello World -->
EOF
sed -i 's/AMPERSAND_TOKEN/\&/g' test1.md
bin/makedo test test1.md
````
<!-- !out Warnings: -->


### Test 2: Background op in the middle

````bash
cat << 'EOF' > test2.md
```bash
echo "Hello World" AMPERSAND_TOKEN echo "ok"
```
<!-- out ok -->
EOF
sed -i 's/AMPERSAND_TOKEN/\&/g' test2.md
bin/makedo test test2.md
````
<!-- !out Warnings: -->


### Test 3: `&` in a string should NOT warn

````bash
cat << 'EOF' > test3.md
```bash
echo "Hello World AMPERSAND_TOKEN"
```
<!-- out Hello World AMPERSAND_TOKEN -->
EOF
sed -i 's/AMPERSAND_TOKEN/\&/g' test3.md
bin/makedo test test3.md
````
<!-- !out Warnings: -->


### Test 4: `&&` logical AND should NOT warn

````bash
cat << 'EOF' > test4.md
```bash
echo "Hello World" && echo Siuum
```
<!-- out Hello World -->
EOF
bin/makedo test test4.md
````
<!-- !out Warnings: -->


### Test 5: Redirection `2>&1` should NOT warn

````bash
cat << 'EOF' > test5.md
```bash
echo Siuuuuum > /dev/null 2>&1
```
<!-- !out Siuuuuum -->
EOF
bin/makedo test test5.md
````
<!-- !out Warnings: -->


### Cleanup

```bash
rm -f test1.md test2.md test3.md test4.md test5.md
```
