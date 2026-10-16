package main

import (
	"makedo/cmd"
	"makedo/internal/version"
)

var versionString string

func main() {
	version.SetVersion(versionString)
	cmd.Execute()
}
