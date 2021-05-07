package manager

import (
	"fmt"
	"net/http"
	"properlyauth/controllers"
	"properlyauth/models"
	"properlyauth/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
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
		userToBeAdded = &models.User{Type: models.Landlord}
		user.Email = data.Email
		names := strings.Split(data.Name, " ")
		if len(names) > 1 {
			user.LastName = names[1]
		}
		user.FirstName = names[0]
		password := utils.GenerateRandom(10)
		user.Password = utils.SHA256Hash(password)
		user.PhoneNumber = data.Phone
		if err := models.Insert(user, models.UserCollectionName); err != nil {
			models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Something went wrong while creating new user"), struct{}{})
			return
		}
		invite := &models.Invite{Type: models.Landlord,
			Email: data.Email,
			Name:  data.Name,
			Phone: data.Phone,
		}

		if err := models.Insert(invite, models.InvitesCollectionName); err != nil {
			models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Something went wrong while creating new user"), struct{}{})
			return
		}
		body = fmt.Sprintf(`
		<h1>You are being invited to join properly as a landlord</h1>
		<p>follow this link to join <a href=%s>sign in</a> properly using the password :%s</p>`, link, password)

	} else {
		if userToBeAdded.Type != models.Landlord {
			models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("User is not a Landlord"), user)
			return
		}
		body = fmt.Sprintf(
			`<h1>You have been Added to property %s by %s</h1>`,
			property.Name, fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		)
	}

	userToBeAdded, err = models.GetUser("email", data.Email)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Fatal error "), struct{}{})
		return
	}
	if err := utils.SendMail(data.Email, "Invitation From Peoperly", body); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}

	property.Landlords[userToBeAdded.ID] = userToBeAdded.ID
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
