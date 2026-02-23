package main

import (
	"embed"

	"go-postfixadmin/cmd"
)

//go:embed views public locales
var embeddedFiles embed.FS

func main() {
	cmd.Execute(embeddedFiles)
}
