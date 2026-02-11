package main

import (
	"embed"
	"flag"
	"fmt"
	"log/slog"
)

const Version = "1.0.0"

//go:embed views public
var embeddedFiles embed.FS

func main() {
	// CLI Flags
	versionFlag := flag.Bool("version", false, "Display version information")
	runFlag := flag.Bool("run", false, "Start the administration server")
	portFlag := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Go-Postfixadmin version %s\n", Version)
		return
	}

	if !*runFlag {
		flag.Usage()
		return
	}

	slog.Info("Starting Go-Postfixadmin...", "version", Version)
	StartServer(embeddedFiles, *portFlag)
}
