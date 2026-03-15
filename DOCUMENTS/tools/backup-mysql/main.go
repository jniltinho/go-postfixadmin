// backup-mysql - MySQL backup tool
// Copyrighted by Nilton OS <jniltinho at gmail.com>
// License: LGPLv3 (http://www.gnu.org/licenses/lgpl.html)
//
// Usage:
//   backup-mysql backup --debug
//   backup-mysql backup --clean=5
//   backup-mysql backup --clean=5 --sendmail
//   backup-mysql list
//
// Crontab example:
//   05 01 * * * /usr/local/bin/backup-mysql backup --clean=5

package main

import (
	"fmt"
	"os"
)

// Injected at build time via: -ldflags "-X 'main.version=x.y.z' -X 'main.buildDate=...'"
var (
	version   = "dev"
	buildDate = "Unknown"
)

func main() {
	cfg := loadConfig()
	if err := newRootCmd(&cfg).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
