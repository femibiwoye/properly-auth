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

	manager := v1.Group("/manager")
	manager.PUT("/create/property/", controllers.CreateProperty)
	manager.PUT("/update/property/", controllers.UpdatePropertyRoute)

	landlord := v1.Group("/landlord")
	landlord.PUT("/property/add/", controllers.AddLandlordToProperty)
	landlord.PUT("/property/remove/", controllers.RemoveLandlordFromProperty)
	landlord.GET("/property/list/", controllers.ListLandlordFromProperty)

	tenant := v1.Group("/tenant")
	tenant.PUT("/property/add/", controllers.AddTenantToProperty)
	tenant.PUT("/property/remove/", controllers.RemoveTenantFromProperty)
	tenant.GET("/property/list/", controllers.ListTenantFromProperty)

	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return app
}
