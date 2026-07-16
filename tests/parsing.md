# Markdown Parsing Edge Cases

This test file verifies that the `makedo` engine correctly parses alternative markdown code block syntaxes. Thanks to our underlying Goldmark parser, we can effortlessly parse advanced fence styles, variable backtick counts, and nested combinations straight out of the box without any manual parsing logic!

## Tilde Fences

Markdown allows code fences to be defined with three or more tildes (`~`).

~~~bash
echo "parsed tilde block"
~~~
<!-- out parsed tilde block -->

## Four Backticks

You can use four backticks to encapsulate blocks. This is especially useful if the block itself contains triple backticks.

````bash
echo "parsed four backticks"
````
<!-- out parsed four backticks -->

## Nested Fences (Backticks inside Tildes)

Tilde fences can safely contain backtick fences without escaping them.

~~~bash
echo '```bash'
echo 'echo hello'
echo '```'
~~~
<!-- outr ```bash\s*echo hello\s*``` -->

## Nested Fences (Tildes inside Backticks)

Similarly, backtick fences can safely contain tilde fences.

```bash
echo "~~~bash"
echo "echo world"
echo "~~~"
```
<!-- outr ~~~bash\s*echo world\s*~~~ -->

## Five Backticks

Any number of backticks three or greater works, as long as the opening and closing fences match.

`````bash
echo "parsed five backticks"
`````
<!-- out parsed five backticks -->
