---
title: "Environment & State"
weight: 4
bookToc: true
---

# Environment & State

MakeDo provides seamless state management by executing all code blocks within a single, persistent shell environment.

## Persistent Shell

Every code block is executed in the **same, persistent shell process**. Variables, aliases, or functions defined in one block persist automatically to the next:

```bash
MY_VAR="defined_here"
echo "MY_VAR is set to: $MY_VAR"
```
<!-- out MY_VAR is set to: defined_here -->

```stdout
MY_VAR is set to: defined_here
```

In the next block, `MY_VAR` is still available:
```bash
echo "MY_VAR is now: ${MY_VAR:-empty}"
```
<!-- out MY_VAR is now: defined_here -->

```stdout
MY_VAR is now: defined_here
```

## Changing Directories

Because the shell is persistent, changing the working directory (`cd`) in one block will carry over to the next blocks automatically:

```bash
mkdir -p /tmp/makedo_doc_env_dir
cd /tmp/makedo_doc_env_dir
```

```bash
pwd
```
<!-- out /tmp/makedo_doc_env_dir -->

```stdout
/tmp/makedo_doc_env_dir
```

## Disk Persistence

Just like variables and directories, filesystem changes persist naturally across blocks:

```bash
echo "persisted" > /tmp/makedo_doc_persist.txt
```

And printing out the content of the file we can see it actually persisted:
```bash
cat /tmp/makedo_doc_persist.txt
```
<!-- out persisted -->

```stdout
persisted
```

## Cleanup

```bash
cd /
rm -rf /tmp/makedo_doc_env_dir /tmp/makedo_doc_persist.txt
```
<!-- !checkpath /tmp/makedo_doc_env_dir -->
<!-- !checkpath /tmp/makedo_doc_persist.txt -->
