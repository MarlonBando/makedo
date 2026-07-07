---
title: "Dynamic Types"
weight: 3
bookToc: true
---

# Dynamic Types

MakeDo supports dynamic type placeholders in `out` and `outr` directives to handle variable output like UUIDs, timestamps, or ephemeral IP addresses.

## Syntax

Use the `${{typeName}}` syntax inside directives:

```html
<!-- out Created user ${{uuid}} on ${{date}} -->
```

## Supported Types

| Type | Pattern | Example Match |
| :--- | :--- | :--- |
| `${{date}}` | `YYYY-MM-DD` | `2024-01-15` |
| `${{time}}` | `HH:MM:SS` | `14:30:00` |
| `${{uuid}}` | Standard UUID | `550e8400-e29b-41d4-a716-446655440000` |
| `${{ip}}` | IPv4 address | `192.168.1.1` |
| `${{number}}` | Integer or decimal | `42`, `3.14` |
| `${{version}}` | Semantic version | `1.2.3`, `1.2.3.4` |

## Examples

### Matching Time

```bash
date +%T
```
<!-- out ${{time}} -->

### Matching UUID

```bash
uuid=$(uuidgen)
echo "$uuid"
```
<!-- out ${{uuid}} -->

### Matching Version

```bash
echo "v3.4.5"
```
<!-- out v${{version}} -->

## How It Works

1. **Pre-compilation**: Before execution, MakeDo scans directives for `${{type}}` placeholders
2. **`outr`**: The placeholder is replaced with its regex pattern before compilation
3. **`out` Upgrade**: If a standard `out` directive contains a placeholder, MakeDo automatically upgrades it to a regex check. Literal parts are properly escaped using `regexp.QuoteMeta`
4. **Embed Substitution**: During `makedo embed`, actual runtime values are replaced back into placeholder syntax to keep embedded output deterministic

## Negation with Types

Dynamic types work with negation too:

```bash
echo "not a uuid"
```
<!-- !outr ${{uuid}} -->
