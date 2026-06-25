# MKENV

Every time we have a fenced code block makedo
will implicitly run `$MAKEDO_ENV` file before the code
inside of it.
So if we want to share variables or setup script between
fenced code block we have to use `$MAKEDO_ENV`
```bash
echo "export FOO=hello" >> $MAKEDO_ENV
```

Here `FOO` has been passed successfully between fenced code block.
```bash
echo "FOO is $FOO"
```
<!-- out FOO is hello -->

Note that if we modify `$MAKEDO_ENV` the changes
will be available from the next fenced code block.
```bash
echo "export FOO=sium" >> $MAKEDO_ENV
echo "FOO is $FOO"
```
<!-- out FOO is hello -->

```bash
echo "FOO is $FOO"
```
<!-- out FOO is sium -->

If we really need instant changes, we can do it manually.
```bash
echo "export FOO=now" >> $MAKEDO_ENV
source $MAKEDO_ENV
echo "FOO is $FOO"
```
<!-- !out FOO is sium -->
<!-- out FOO is now -->

We can also use it for command, let's say we want to know change folder.

```bash
mkdir f
echo "hello world" >> f/file.txt
echo "cd f" >> $MAKEDO_ENV
```
<!-- checkpath f/file.txt -->

Ok now cat file.txt should prin hello world

```bash
cat file.txt
```
<!-- out hello world -->

```bash
rm file.txt
cd ..
rmdir f
```
<!-- !checkpath f/file.txt -->

## CMD Env Sourcing

Let's check if environment variables from `$MAKEDO_ENV` are available to cmd directives.

```bash
echo "CMD_TEST_VAR=cmd_success" >> $MAKEDO_ENV
```

```bash
echo "Verifying CMD_TEST_VAR in cmd directive..."
```
`<!-- cmd [ "$CMD_TEST_VAR" = "cmd_success" ] -->`
