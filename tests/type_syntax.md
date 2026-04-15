# Type Syntax Test

```bash
echo "Current Date: 2023-10-25"
```
<!-- outr Current Date: \d{4}-\d{2}-\d{2} -->
<!-- out Current Date: ${{date}} -->

```bash
echo '{"id":"550e8400-e29b-41d4-a716-446655440000"}'
```
<!-- out {"id":"${{uuid}}"} -->
