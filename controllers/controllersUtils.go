package controllers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"properlyauth/models"
	"properlyauth/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
)

type managerRequestData interface {
	GetUserID() string
	GetPropertyID() string
}

//UpdateUser update a user data in the DB
func UpdateData(data models.ProperlyDocModel, collectionName string) error {
	uB, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	var update bson.M
	err = bson.Unmarshal(uB, &update)
	if err != nil {
		return err
	}
	err = models.Update(data, bson.D{{Key: "$set", Value: update}}, collectionName)
	if err != nil {
		return err
	}

	return nil
}

//HandleMediaUploads helper function to upload media files
func HandleMediaUploads(c *gin.Context, nameOf string, form *multipart.Form) ([]string, error) {
	files := form.File[nameOf]
	names := []string{}
	rootDir := os.Getenv("ROOTDIR")
	errors := []error{}
	for _, file := range files {
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Base(file.Filename))
		err := c.SaveUploadedFile(file, fmt.Sprintf("%s/public/media/%s", rootDir, filename))
		if err != nil {
			errors = append(errors, err)
			continue
		}
		names = append(names, filename)
	}
	if len(errors) > 0 {
		return names, errors[0]
	}
	return names, nil
}

//CheckUser helper function to validate a user and optional check if he is manager
func CheckUser(c *gin.Context, checkManager bool) (*models.User, string, bool) {
	platform, err := GetPlatform(c)
	if err != nil {
		return nil, "", false
	}
	res, err := utils.DecodeJWT(c)
	if err != nil {
		models.NewResponse(c, http.StatusUnauthorized, err, struct{}{})
		return nil, "", false
	}

	userM, err := models.FetchDocByCriterion("id", res["user_id"], models.UserCollectionName)

	if userM == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("user not found"), struct{}{})
		return nil, "", false
	}
	userFetch, err := models.ToUserFromM(userM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return nil, "", false
	}
	if checkManager {
		if userFetch.Type != models.Manager {
			models.NewResponse(c, http.StatusUnauthorized, fmt.Errorf("Only managers can create and change properties"), userFetch)
			return nil, "", false
		}
	}

	return userFetch, platform, true
}

//ValidateProperty validate a give property id and user
func ValidateProperty(c *gin.Context,
	data managerRequestData,
	checkUserID bool, // used to confirm if a user id should be check in the request
	typed, // use to indicate what type of operation we are operating
	operation string) (
	*models.Property,
	*models.User, // The user to change it's details
	*models.User, // The user making the request
	bool,
) {
	user, _, ok := CheckUser(c, true)
	if !ok {
		return nil, nil, nil, ok
	}
	c.ShouldBindJSON(&data)

	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return nil, nil, nil, false
	}

	// we can omit user ID for
	if !checkUserID {
		delete(errorResponse, "UserID")
	}
	if len(errorResponse) > 0 {
		_, ok := errorResponse["UserID"]
		if !ok && operation != "List" {
			models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided invalid %s property details", operation), errorResponse)
			return nil, nil, nil, false
		}
	}

	propertyM, _ := models.FetchDocByCriterion("id", data.GetPropertyID(), models.PropertyCollectionName)
	if propertyM == nil {
		_, ok := errorResponse["PropertyID"]
		if !ok {
			errorResponse["propertyid"] = []string{"Property id doesn't match  any property"}
		}
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Property not found"), errorResponse)
		return nil, nil, nil, false
	}
	property, err := models.ToPropertyFromM(propertyM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return nil, nil, nil, false
	}

	userM, err := models.FetchDocByCriterion("id", data.GetUserID(), models.UserCollectionName)

	if userM == nil && checkUserID {
		errorResponse["userid"] = []string{"User id doesn't match  any user"}
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("The User to %s not found", operation), errorResponse)
		return nil, nil, nil, false
	}

	userFetch, err := models.ToUserFromM(userM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return nil, nil, nil, false
	}

	if checkUserID && userFetch.Type != typed {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Can't not %s non %s to property using this endpoint", operation, typed), struct{}{})
		return nil, nil, nil, false
	}

	return property, userFetch, user, true
}

//AugmentProperty helper function to change sub data of a property
func AugmentProperty(c *gin.Context, typed, operation string, f func(map[string]string, string)) {
	data := models.AddLandlord{}
	property, userFetch, _, ok := ValidateProperty(c, &data, true, typed, operation)

	if !ok {
		return
	}
	field := "id"
	values := []string{}
	switch typed {
	case models.Landlord:
		f(property.Landlords, userFetch.ID)
		values = mapKeysToArray(property.Landlords)
	case models.Tenant:
		f(property.Tenants, userFetch.ID)
		values = mapKeysToArray(property.Tenants)
	case models.Vendor:
		f(property.Vendors, userFetch.ID)
		values = mapKeysToArray(property.Vendors)
	}

	users, err := models.FetchDocByCriterionMultiple(field, models.UserCollectionName, values)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	UpdateData(property, models.PropertyCollectionName)

	models.NewResponse(c, http.StatusOK, fmt.Errorf("%s %s from this property", typed, operation), users)

}

func mapKeysToArray(m map[string]string) []string {
	results := []string{}
	for key := range m {
		results = append(results, key)
	}
	return results
}

func FetchList(c *gin.Context, typed string) {
	data := models.AddLandlord{}
	property, _, _, ok := ValidateProperty(c, &data, false, typed, "List")
	if !ok {
		return
	}

	field := "id"
	values := []string{}
	switch typed {
	case models.Landlord:
		values = mapKeysToArray(property.Landlords)
	case models.Tenant:
		values = mapKeysToArray(property.Tenants)
	case models.Vendor:
		values = mapKeysToArray(property.Vendors)
	}

	users, err := models.FetchDocByCriterionMultiple(field, models.UserCollectionName, values)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("List of %s added to this property", typed), users)
}

//GetPlatform get the platform type for this request
func GetPlatform(c *gin.Context) (string, error) {
	query := c.Request.URL.Query()
	platform, ok := query["platform"]

	if !ok || len(platform) <= 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid platform"), nil)
		return "", fmt.Errorf("No query sent for platform type sent")
	}
	return strings.Trim(platform[0], " "), nil
}

//ErrorReponses generate error response for empty data
func ErrorReponses(c *gin.Context, data interface{}, api string) (string, bool) {
	platform, err := GetPlatform(c)
	if err != nil {
		return platform, true
	}

	c.ShouldBindJSON(&data)
	errorReponse, err := utils.MissingDataResponse(data)
	if len(errorReponse) > 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided invalid %s details", api), errorReponse)
		return platform, true
	}
	return platform, false
}
