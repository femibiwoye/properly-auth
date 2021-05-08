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

// AddTenantToProperty godoc
// @Summary endpoint to add a tenant to a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.AddTenantProperty true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /tenant/property/add/ [post]
// @Security ApiKeyAuth
func AddTenantToProperty(c *gin.Context) {
	user, _, ok := controllers.CheckUser(c, true)
	if !ok {
		return
	}
	data := models.AddTenantProperty{}
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
	if !utils.CheckDateFormat(data.RentStartDate) || !utils.CheckDateFormat(data.RentEndDate) {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid date"), struct{}{})
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
		userToBeAdded = &models.User{Type: models.Tenant, PhoneNumber: data.Phone, Email: data.Email}
		names := strings.Split(data.Name, " ")
		if len(names) > 1 {
			userToBeAdded.LastName = names[1]
		}
		userToBeAdded.FirstName = names[0]
		password := utils.GenerateRandom(10)
		userToBeAdded.Password = utils.SHA256Hash(password)
		if err := models.Insert(userToBeAdded, models.UserCollectionName); err != nil {
			models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Something went wrong while creating new user"), err.Error())
			return
		}
		invite := &models.Invite{Type: models.Landlord,
			Email: data.Email,
			Name:  data.Name,
			Phone: data.Phone,
		}
		if err := models.Upsert(invite, map[string]interface{}{"email": data.Email}, models.InvitesCollectionName); err != nil {
			models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Something went wrong while creating new user"), err.Error())
			return
		}
		body = fmt.Sprintf(`
		<h1>You are being invited to join properly as a tenant</h1>
		<p>follow this link to join %s or use this password to login with you mail password :%s</p>`, link, password)

	} else {
		if userToBeAdded.Type != models.Tenant {
			models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("User is not a Tenant"), struct{}{})
			return
		}
		body = fmt.Sprintf(
			`<h1>You have been Added to property %s by %s</h1>`,
			property.Name, fmt.Sprintf("%s %s", property.Name, user.LastName),
		)
	}

	userToBeAdded, err = models.GetUser("email", data.Email)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Fatal error "), err)
		return
	}
	if err := utils.SendMail(data.Email, "Invitation From Peoperly", body); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error sending mail"), err)
		return
	}
	property.Tenants[userToBeAdded.ID] = fmt.Sprintf("%s %s", userToBeAdded.FirstName, userToBeAdded.LastName)
	if err = controllers.UpdateData(property, models.PropertyCollectionName); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}
	models.NewResponse(c, http.StatusOK, fmt.Errorf("User added to property"), property)
}

// RemoveTenantFromProperty godoc
// @Summary endpoint to remove a tanent from a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.AugmentProperty true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /tenant/property/remove/ [delete]
// @Security ApiKeyAuth
func RemoveTenantFromProperty(c *gin.Context) {
	controllers.AugmentProperty(c, models.Tenant, "remove", func(m map[string]string, id string) {
		delete(m, id)
	})
}

// ListTenantFromProperty godoc
// @Summary list all tenant in a property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.ListType true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /tenant/property/list/ [get]
// @Security ApiKeyAuth
func ListTenantFromProperty(c *gin.Context) {
	controllers.FetchList(c, models.Tenant)
}
