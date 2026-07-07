---
title: "Environment & State"
weight: 4
bookToc: true
---

# Environment & State

MakeDo provides mechanisms for managing state across code blocks despite each block running in an isolated shell.

## Shell Isolation

Every code block is executed in a **new, independent shell process**. Variables or functions defined in one block do **not** persist to the next:

```bash
MY_VAR="defined_here"
echo "MY_VAR is set to: $MY_VAR"
```
<!-- out MY_VAR is set to: defined_here -->

```stdout
MY_VAR is set to: defined_here
```

In the next block MY_VAR is not defined
```bash
echo "MY_VAR is now: ${MY_VAR:-empty}"
```
<!-- out MY_VAR is now: empty -->

```stdout
MY_VAR is now: empty
```

## Sharing State with `$MAKEDO_ENV`

To share variables between blocks, append exports to the `$MAKEDO_ENV` file. This file is automatically sourced at the beginning of all subsequent code blocks:

```bash
echo "export SHARED_VAR=hello_world" >> $MAKEDO_ENV
```

Now `SHARED_VAR` is available in the next block:

```bash
echo "SHARED_VAR is $SHARED_VAR"
```
<!-- out SHARED_VAR is hello_world -->

```stdout
SHARED_VAR is hello_world
```

{{< hint info >}}
**Note:** Changes to `$MAKEDO_ENV` take effect in the *next* code block, not the current one.
{{< /hint >}}

## Changing Directories

You can also use `$MAKEDO_ENV` to change the working directory for subsequent blocks:

```bash
mkdir -p /tmp/makedo_doc_env_dir
echo "cd /tmp/makedo_doc_env_dir" >> $MAKEDO_ENV
```

```bash
pwd
```
<!-- /tmp/makedo_doc_env_dir -->

```stdout
/tmp/makedo_doc_env_dir
```

## Instant Changes

If you need changes immediately in the same block, source `$MAKEDO_ENV` manually:

```bash
echo "export INSTANT_VAR=now" >> $MAKEDO_ENV
source $MAKEDO_ENV
echo "INSTANT_VAR is $INSTANT_VAR"
```
<!-- out INSTANT_VAR is now -->

```stdout
INSTANT_VAR is now
```

## Disk Persistence

While environment variables require `$MAKEDO_ENV`, filesystem changes persist across blocks automatically:

```bash
echo "persisted" > /tmp/makedo_doc_persist.txt
```

And printing out the content of the file we can see it acutally persisted
```bash
cat /tmp/makedo_doc_persist.txt
```
<!-- out persisted -->

```stdout
persisted
```

## Cleanup

```bash
rm -rf /tmp/makedo_doc_env_dir /tmp/makedo_doc_persist.txt
```
<!-- !checkpath /tmp/makedo_doc_env_dir -->
<!-- !checkpath /tmp/makedo_doc_persist.txt -->
