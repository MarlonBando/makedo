# Type Syntax Test

This is how you get the time in bash
```bash
date +%T
```
<!-- out ${{time}} -->
```stdout
${{time}}
```

This is a random uuid
```bash
uuid=$(uuidgen)
echo "$uuid"   
```
<!-- out ${{uuid}} -->
```stdout
${{uuid}}
```
```bash
echo "v3.4.5"
```
<!-- out v${{version}} -->
```stdout
v${{version}}
```
