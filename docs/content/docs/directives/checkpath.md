---
title: "checkpath"
weight: 4
bookToc: true
---

# The `checkpath` Directive

The `checkpath` directive asserts that a file or directory exists on the filesystem.

## Check File Exists

```bash
touch /tmp/makedo_doc_checkpath.txt
```
<!-- checkpath /tmp/makedo_doc_checkpath.txt -->

## Check Directory Exists

```bash
mkdir -p /tmp/makedo_doc_checkdir
```
<!-- checkpath /tmp/makedo_doc_checkdir -->

## Negation

Use `!checkpath` to assert a path does NOT exist:

```bash
echo "checking non-existent path"
```
<!-- !checkpath /tmp/definitely_does_not_exist_makedo_99.txt -->

## Cleanup

```bash
rm -f /tmp/makedo_doc_checkpath.txt
rm -rf /tmp/makedo_doc_checkdir
```
<!-- !checkpath /tmp/makedo_doc_checkpath.txt -->
<!-- !checkpath /tmp/makedo_doc_checkdir -->
