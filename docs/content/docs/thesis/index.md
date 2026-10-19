---
title: "Thesis"
weight: 8
bookToc: true
---

# MakeDo: From README to Test Suite

MakeDo was designed and evaluated in an MSc thesis at the Technical University of Denmark (DTU), *MakeDo: From README to Test Suite. An Execution Model for Shell Blocks in Markdown* (Michele Bandini, 2026).

{{< button href="makedo-thesis.pdf" >}}Download the thesis (PDF){{< /button >}}

## Abstract

Almost every repository contains a README, and many use it to document how to run a CLI tool. As the software changes, the commands in those Markdown files stop behaving as documented: this is documentation drift. The thesis introduces MakeDo, a Markdown annotation system that makes Markdown files act as complete tests, so a failing test means the documentation has drifted.

## Findings

- **Docs break and stay broken.** Over 9 years, Hugo's quick-start guide broke 6 times and stayed broken for 798 days in total. One breakage, caused by an unpinned theme, lasted 520 days and 905 commits. In a survey of 62 popular repositories, 79% have shell examples in their docs and 67% of those never test them.
- **MakeDo finds real bugs.** Changing 19 code fences in ripgrep's `GUIDE.md` to `bash` and running `makedo test` once revealed 6 broken examples nobody had reported ([ripgrep#3448](https://github.com/BurntSushi/ripgrep/issues/3448)). Replaying the commit that broke them in CI showed MakeDo passing before it and failing after it.
- **It works as a test framework.** 70 of the 75 tests in Hugo's integration suite were ported to MakeDo. The port caught the same 8 of 10 injected bugs as Hugo's original `testscript` suite, with 12 lines of Bash helpers instead of 352 lines of Go.

## Contents

1. **Background:** documentation drift, literate programming and testing, and existing executable-Markdown tools.
2. **MakeDo:** design goals, the document and execution model, block lifecycle, directive and embed semantics.
3. **Evaluation:** the Hugo, ripgrep and Hugo integration-suite case studies.
4. **Conclusion:** limitations and future work.

## Cite

```bibtex
@mastersthesis{bandini2026makedo,
  author = {Michele Bandini},
  title  = {MakeDo: From README to Test Suite. An Execution Model for Shell Blocks in Markdown},
  school = {Technical University of Denmark},
  year   = {2026}
}
```
