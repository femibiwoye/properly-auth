package main

import (
	"log"
	"os"
	"properlyauth/routes"

	"github.com/joho/godotenv"

	"properlyauth/docs"
)

// @title Swagger Example API
// @version 1.0
// @description properly backend.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.implicit OAuth2Implicit
// @authorizationUrl https://example.com/oauth/authorize
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.password OAuth2Password
// @tokenUrl https://example.com/oauth/token
// @scope.read Grants read access
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.accessCode OAuth2AccessCode
// @tokenUrl https://example.com/oauth/token
// @authorizationUrl https://example.com/oauth/authorize
// @scope.admin Grants read and write access to administrative information
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	router, chatServer := routes.Router()
	defer chatServer.Close()

	docs.SwaggerInfo.Title = "Properly Swagger Docs"
	docs.SwaggerInfo.Description = "This is a  Properly  API."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = os.Getenv("HOST")
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	router.Static("/public", "public")
	router.Run()
}
