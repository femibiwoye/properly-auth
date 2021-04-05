package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"properlyauth/controllers"
	genaralRoutes "properlyauth/controllers/general"
	managerRoutes "properlyauth/controllers/manager"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// ServeFile godoc
// @Summary endpoints to return save files
// @Description
// @Tags accounts
// @Accept  json
// @Produce  gif,png,jpeg
// @Router /serve/media/:filename [get]
// @Security ApiKeyAuth
func ServeFile(c *gin.Context) {
	name := c.Param("filename")
	rootDir := os.Getenv("ROOTDIR")
	c.File(fmt.Sprintf("%s/public/media/%s", rootDir, name))
}

//Router instanciate all routes in the application
func Router() *gin.Engine {

	app := gin.Default()

	v1 := app.Group("/v1")

	v1.GET("/", func(c *gin.Context) {
		c.String(200, "Welcome to properly")
	})
	v1.GET("/serve/media/:filename", ServeFile)

	v1.POST("/signup/", controllers.SignUp)
	v1.PUT("/reset/update-password/", controllers.ResetPassword)
	v1.PUT("/user/change-password/", controllers.ChangePasswordAuth)
	v1.POST("/reset/validate-token/", controllers.ChangePasswordFromToken)
	v1.POST("/login/", controllers.SignIn)
	v1.GET("/user/", controllers.UserProfile)

	v1.PUT("/user/update/", controllers.UpdateProfile)
	v1.PUT("/user/update-profile-image/", controllers.UpdateProfileImage)

	v1.POST("/make/complaint/", genaralRoutes.MakeComplaints)
	v1.PUT("/update/complaint/", genaralRoutes.UpdateComplaints)
	v1.GET("/list/complaints/", genaralRoutes.ListComplaints)

	manager := v1.Group("/manager")
	manager.PUT("/create/property/", managerRoutes.CreateProperty)
	manager.PUT("/update/property/", managerRoutes.UpdatePropertyRoute)
	manager.DELETE("/remove/attachment/", managerRoutes.RemoveAttachment)
	manager.POST("/inspection/schedule/", managerRoutes.ScheduleInspection)
	manager.PUT("/inspection/update/", managerRoutes.UpdateInspection)
	manager.DELETE("/inspection/delete/", managerRoutes.DeleteInspection)
	manager.GET("/list/properties/", managerRoutes.ListProperties)


	landlord := v1.Group("/landlord")
	landlord.POST("/property/add/", managerRoutes.AddLandlordToProperty)
	landlord.DELETE("/property/remove/", managerRoutes.RemoveLandlordFromProperty)
	landlord.GET("/property/list/", managerRoutes.ListLandlordFromProperty)

	tenant := v1.Group("/tenant")
	tenant.POST("/property/add/", managerRoutes.AddTenantToProperty)
	tenant.DELETE("/property/remove/", managerRoutes.RemoveTenantFromProperty)
	tenant.GET("/property/list/", managerRoutes.ListTenantFromProperty)

	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return app
}
