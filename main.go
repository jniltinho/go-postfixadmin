package main

import (
	"embed"

	_ "go-postfixadmin/docs" // swagger docs (must be imported so init() registers the spec)

	"go-postfixadmin/cmd"
)

// @title Go-PostfixAdmin API
// @version 1.0
// @description Modern REST API for Go-PostfixAdmin with JWT authentication.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/jniltinho/go-postfixadmin
// @contact.email jniltinho@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

//go:embed web/dist web/files all:locales
// web/files/config.default.toml is embedded and used by the --generate-config command
var embeddedFiles embed.FS

func main() {
	cmd.Execute(embeddedFiles)
}
