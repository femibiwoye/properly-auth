package manager

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"properlyauth/controllers"
	"properlyauth/models"
	"properlyauth/utils"
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
	user, _, ok := controllers.CheckUser(c, true)
	if !ok {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Not authorize"), false)
		return
	}
	data := models.AddLandLordProperty{}
	c.ShouldBindJSON(&data)
	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}
	if len(errorResponse) > 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided incomplete requests details"), errorResponse)
		return
	}
	userToBeAdded, err := models.GetUser("email", data.Email)
	if err != nil {
		if err.Error() != mongo.ErrNoDocuments.Error() {
			models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Email: "), errorResponse)
			return
		}
	}

	property, err := models.GetProperty("id", data.PropertyID)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Property Not found "), errorResponse)
		return
	}
	body := ``
	link := ""
	if userToBeAdded == nil {
		//send mail to user to register
		body = fmt.Sprintf(`
		<h1>You are being invited to join properly as a landlord</h1>
		<p>follow this link to join %s</p>`, link)

	} else {
		body = fmt.Sprintf(
			`<h1>You have been Added to a property %s by %s</h1>`,
			property.Name, fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		)
	}

	if err := utils.SendMail(data.Email, "Invitation From Peoperly", body); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}

	if userToBeAdded ==nil{
		models.NewResponse(c, http.StatusOK, fmt.Errorf("User not yet registered. An Email Invite has been sent to the person"), struct{}{})
		return
	}

	property.Landlords[data.UserID] = data.UserID
	if err = controllers.UpdateData(property, models.PropertyCollectionName); err != nil {
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
// @Router /landlord/property/list/ [get]
// @Security ApiKeyAuth
func ListLandlordFromProperty(c *gin.Context) {
	controllers.FetchList(c, models.Landlord)
}
