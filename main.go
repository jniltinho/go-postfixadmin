package main

import (
	"embed"

	"go-postfixadmin/cmd"
)

//go:embed web locales
var embeddedFiles embed.FS

func main() {
	cmd.Execute(embeddedFiles)
}
