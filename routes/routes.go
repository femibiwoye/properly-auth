package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"properlyauth/controllers"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//Router instanciate all routes in the application
func Router() *gin.Engine {

	app := gin.Default()

	app.GET("/", func(c *gin.Context) {
		c.String(200, "Welcome to properly")
	})
	app.GET("/serve/media/:filename", func(c *gin.Context) {
		name := c.Param("filename")
		rootDir := os.Getenv("ROOTDIR")
		c.File(fmt.Sprintf("%s/public/media/%s", rootDir, name))
		//TODO autheticate this url
	})

	app.POST("/signup/", controllers.SignUp)
	app.POST("/reset/password/", controllers.ResetPassword)
	app.POST("/change/password/auth/", controllers.ChangePasswordAuth)
	app.POST("/change/password/token/", controllers.ChangePasswordFromToken)
	app.POST("/signin/", controllers.SignIn)
	app.GET("/generate/pumc/", controllers.GeneratePUMC)
	app.GET("/profile/", controllers.UserProfile)

	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return app
}
