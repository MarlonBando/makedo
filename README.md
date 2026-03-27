# makedo

Makedo is a tool that bring integration testing directly inside your markdown documentation.
Documentation becomes the tests and tests become doumentation.

Makedo allows you to:
- Verify that your documentation is still saying the truth
- Run integration testing on your cli tool using markdown
- Run all the codes in a markdown file (useful for setup)

## How does it work?
Makedo looks for fenced code block followed by html comment in this format <!-- keyword content -->
The content of the comment tells makedo what expect from the code block output.
When markdown is rendered makedo directives will be hidden, to see them you have to view the raw code

Let's say we want to verify the flag `-h` of makedo does its job correctly and it prints how to use it.
We have to add `<!-- out makedo  -->` after the fenced block. We are saying, check that the output contains `makedo [command]`

```bash
makedo -h
```
 <!-- out A markdown-based task runner -->

We can also run a cli command to validate our fenced block.
For instance our fenced block just installed Git, manually we would use `git -v` to check if it's installed.
We can tell makedo to do so with `<!-- cmd git -v -->`

```bash
echo "Pretending to install git..."
```
<!-- cmd git -v -->

### Available directives
At the moment we have the following directives:
- out  [content] -> check if stout contains [content]
- outr [regex] -> check if stout contains [regex]
- cmd  [content] -> run a command. If it succed the test is passed
- pwd  [content] -> check if pwd print [content]
