package manager

import (
	"github.com/gin-gonic/gin"
	"properlyauth/controllers"
	"properlyauth/models"
)

// AddLandlordToProperty godoc
// @Summary endpoint to add a landloard to a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.AddLandlord true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /landlord/property/add/ [post]
// @Security ApiKeyAuth
func AddLandlordToProperty(c *gin.Context) {
	controllers.AugmentProperty(c, models.Landlord, "add", func(m map[string]string, id string) {
		m[id] = id
	})
}

// RemoveLandlordFromProperty godoc
// @Summary endpoint to remove a landloard from a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.AddLandlord true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /landlord/property/remove/ [delete]
// @Security ApiKeyAuth
func RemoveLandlordFromProperty(c *gin.Context) {
	controllers.AugmentProperty(c, models.Landlord, "remove", func(m map[string]string, id string) {
		delete(m, id)
	})
}

// ListLandlordFromProperty godoc
// @Summary return the landlords added to a property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.ListType true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /landlord/property/list/ [get]
// @Security ApiKeyAuth
func ListLandlordFromProperty(c *gin.Context) {
	controllers.FetchList(c, models.Landlord)
}
