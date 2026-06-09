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
