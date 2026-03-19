package main

import (
	"fmt"
	"github.com/yuin/goldmark"
	// "github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"os"
	"strings"
)

func main() {
	// /home/bando/projects/thesis/hugo/docs/content/en/getting-started

	mdFile := "/home/bando/projects/thesis/hugo/docs/content/en/getting-started/quick-start.md"
	source, err := os.ReadFile(mdFile)
	if err != nil {
		panic(err)
	}

	md := goldmark.New()
	reader := text.NewReader(source)

	doc := md.Parser().Parse(reader)

	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		return getCodeBlocks(source, n, entering)
	})
}

func getCodeBlocks(source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	fmt.Println(n.Kind())
	switch n.Kind() {
	case ast.KindFencedCodeBlock:
		{
			// We move from interface to specific node type
			codeNode, _ := n.(*ast.FencedCodeBlock)

			var code strings.Builder
			// 2. Extract the actual code lines
			for i := 0; i < codeNode.Lines().Len(); i++ {
				code.Reset()
				line := codeNode.Lines().At(i)
				code.Write(line.Value(source))
				fmt.Print(code.String())
			}

		}
	case ast.KindCodeBlock:
		{
			return ast.WalkContinue, nil
		}
	}

	return ast.WalkContinue, nil
}
