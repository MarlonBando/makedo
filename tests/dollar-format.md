# Dollar format 

Dollar format is the following

```bash
$ echo "Hello world"
```
<!-- out Hello world -->

When the stdout is embedded in dollar format it is displayed under the
command inside the fenced code block

```bash
$ echo "Hello world"
Hello world 
```
<!-- out Hello world -->


Multiple lines? No problem.



```bash
$ echo "Hello world"
$ echo "Hello world x2"
Hello world 
```
<!-- out Hello world -->


The dollar format works if we see a dollar as first non white character
we consider it dollar format.
All the following line with `$` are part of the code, 
as soon as a line doesn't start with `$` we consider it part of the output.

```bash
$ echo "Hello world"
Hello world
$ echo "This is not executed"
```
<!-- out Hello world -->
<!-- !out This is not executed -->
