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

	v1 := app.Group("/v1")

	v1.GET("/", func(c *gin.Context) {
		c.String(200, "Welcome to properly")
	})
	v1.GET("/serve/media/:filename", func(c *gin.Context) {
		name := c.Param("filename")
		rootDir := os.Getenv("ROOTDIR")
		c.File(fmt.Sprintf("%s/public/media/%s", rootDir, name))
	})

	v1.POST("/signup/", controllers.SignUp)
	v1.PUT("/reset/update-password/", controllers.ResetPassword)
	v1.PUT("/user/change-password/", controllers.ChangePasswordAuth)
	v1.POST("/reset/validate-token/", controllers.ChangePasswordFromToken)
	v1.POST("/login/", controllers.SignIn)
	v1.GET("/user/", controllers.UserProfile)

	v1.PUT("/user/update/", controllers.UpdateProfile)
	v1.PUT("/user/update-profile-image/", controllers.UpdateProfileImage)

	v1.PUT("/create/property/", controllers.CreateProperty)
	v1.PUT("/update/property/", controllers.UpdatePropertyRoute)

	v1.PUT("/property/add-landlord/", controllers.AddLandlordToProperty)
	v1.PUT("/property/remove-landlord/", controllers.RemoveLandlordFromProperty)
	v1.PUT("/property/add-tenant/", controllers.AddTenantToProperty)
	v1.PUT("/property/remove-tenant/", controllers.RemoveTenantFromProperty)

	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return app
}
