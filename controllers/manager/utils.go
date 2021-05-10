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

func getAddToPropertyDetails(c *gin.Context, accountType string) (
	*models.User, //user adding another user
	*models.User, // user to be added
	*models.Property, // the property
	models.AddToProperty,
	bool, // if all went well
) {
	user, _, ok := controllers.CheckUser(c, true)
	if !ok {
		return nil, nil, nil, nil, false
	}
	var data models.AddToProperty
	if accountType == models.Tenant {
		data = &models.AddTenantProperty{}
	} else {
		data = &models.AddLandLordProperty{}
	}
	c.ShouldBindJSON(&data)
	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return nil, nil, nil, nil, false
	}
	if len(errorResponse) > 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided incomplete requests details"), errorResponse)
		return nil, nil, nil, nil, false
	}
	userToBeAdded, err := models.GetUser("email", data.GetEmail())
	if err != nil {
		if err.Error() != mongo.ErrNoDocuments.Error() {
			models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Email: "), err)
			return nil, nil, nil, nil, false
		}
	}
	property, err := models.GetProperty("id", data.GetPropertyID())
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Property Not found "), errorResponse)
		return nil, nil, nil, nil, false
	}

	return user, userToBeAdded, property, data, true
}

func sendMailToAddedUser(
	c *gin.Context,
	user *models.User,
	userToBeAdded *models.User,
	property *models.Property,
	data models.AddToProperty,
	accountType string,
) (*models.User, bool) {
	body := ``
	link := ""
	if userToBeAdded == nil {
		//send mail to user to register
		userToBeAdded = &models.User{Type: models.Tenant, PhoneNumber: data.GetName(), Email: data.GetEmail()}
		names := strings.Split(data.GetName(), " ")
		if len(names) > 1 {
			userToBeAdded.LastName = names[1]
		}
		userToBeAdded.FirstName = names[0]
		password := utils.GenerateRandom(10)
		userToBeAdded.Password = utils.SHA256Hash(password)
		if err := models.Insert(userToBeAdded, models.UserCollectionName); err != nil {
			models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Something went wrong while creating new user"), err.Error())
			return userToBeAdded, false
		}
		invite := &models.Invite{Type: models.Landlord,
			Email: data.GetEmail(),
			Name:  data.GetName(),
			Phone: data.GetPhone(),
		}
		if err := models.Upsert(invite, map[string]interface{}{"email": data.GetEmail()}, models.InvitesCollectionName); err != nil {
			models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Something went wrong while creating new user"), err.Error())
			return userToBeAdded, false
		}
		body = fmt.Sprintf(`
		<h1>You are being invited to join properly as a %s</h1>
		<p>follow this link to join %s or use this password to login with you mail password :%s</p>`, accountType, link, password)

	} else {
		if userToBeAdded.Type != accountType {
			models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("User is not a %s", accountType), struct{}{})
			return userToBeAdded, false
		}
		body = fmt.Sprintf(
			`<h1>You have been Added to property %s by %s</h1>`,
			property.Name, fmt.Sprintf("%s %s", property.Name, user.LastName),
		)
	}

	userToBeAdded, err := models.GetUser("email", data.GetEmail())
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Fatal error "), err)
		return userToBeAdded, false
	}
	if err := utils.SendMail(data.GetEmail(), "Invitation From Properly", body); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error sending mail"), err)
		return userToBeAdded, false
	}

	return userToBeAdded, true
}
