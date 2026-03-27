package main

import (
	_ "embed"
	"makedo/cmd"
	"makedo/internal/version"
)

//go:embed VERSION
var versionString string

func main() {
	version.SetVersion(versionString)
	cmd.Execute()
}
