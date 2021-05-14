package manager

import (
	"fmt"
	"net/http"
	"properlyauth/controllers"
	"properlyauth/models"
	"properlyauth/utils"

	"github.com/gin-gonic/gin"
)

// AddLandlordToProperty godoc
// @Summary endpoint to add a landloard to a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.AddLandLordProperty true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /landlord/property/add/ [post]
// @Security ApiKeyAuth
func AddLandlordToProperty(c *gin.Context) {
	user, userToBeAdded, property, data, ok := getAddToPropertyDetails(c, models.Landlord)
	if !ok {
		return
	}
	userToBeAdded, ok = sendMailToAddedUser(c, user, userToBeAdded, property, data, models.Landlord)
	if !ok {
		return
	}
	utils.PrintSomeThing(property, userToBeAdded)
	property.Landlords[userToBeAdded.ID] = fmt.Sprintf("%s %s", userToBeAdded.FirstName, userToBeAdded.LastName)
	if err := controllers.UpdateData(property, models.PropertyCollectionName); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}
	models.NewResponse(c, http.StatusOK, fmt.Errorf("User added to property"), property)
}

// RemoveLandlordFromProperty godoc
// @Summary endpoint to remove a landloard from a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.AugmentProperty true "useraccountdetails"
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
// @Router /landlord/property/list/ [post]
// @Security ApiKeyAuth
func ListLandlordFromProperty(c *gin.Context) {
	controllers.FetchList(c, models.Landlord)
}
