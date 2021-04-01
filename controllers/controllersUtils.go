package controllers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"properlyauth/models"
	"properlyauth/utils"
	"time"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
)

func updateProperty(property *models.Property) error {
	uB, err := bson.Marshal(property)
	if err != nil {
		return err
	}
	var update bson.M
	err = bson.Unmarshal(uB, &update)
	if err != nil {
		return err
	}
	err = models.UpdateProperty(property, bson.D{{Key: "$set", Value: update}})
	if err != nil {
		return err
	}

	return nil
}

func handleMediaUploads(c *gin.Context, nameOf string, form *multipart.Form) ([]string, error) {
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

func checkUser(c *gin.Context, checkManager bool) (*models.User, string, bool) {
	platform, err := getPlatform(c)
	if err != nil {
		return nil, "", false
	}
	res, err := utils.DecodeJWT(c)
	if err != nil {
		models.NewResponse(c, http.StatusUnauthorized, err, nil)
		return nil, "", false
	}

	userFetch, _ := models.FetchUserByCriterion("id", res["user_id"])

	if userFetch == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Token doesn't match any user"), struct{}{})
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

func validateProperty(c *gin.Context, typed, operation string) (*models.Property, *models.User, bool) {
	_, _, ok := checkUser(c, true)
	if !ok {
		return nil, nil, ok
	}

	data := models.AddLandlord{}
	c.ShouldBindJSON(&data)

	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return nil, nil, false
	}

	if operation == "List" {
		delete(errorResponse, "UserID")
	}
	if len(errorResponse) > 0 {
		_, ok := errorResponse["UserID"]
		if !ok && operation != "List" {
			models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided invalid %s property details", operation), errorResponse)
			return nil, nil, false
		}
	}

	property, _ := models.FetchPropertyByCriterion("id", data.PropertyID)
	if property == nil {
		_, ok := errorResponse["PropertyID"]
		if !ok {
			errorResponse["propertyid"] = []string{"Property id doesn't match  any property"}
		}
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Property not found"), errorResponse)
		return nil, nil, false
	}

	userFetch, _ := models.FetchUserByCriterion("id", data.UserID)

	if userFetch == nil && operation != "List" {
		errorResponse["userid"] = []string{"User id doesn't match  any user"}
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("The User to %s not found", operation), errorResponse)
		return nil, nil, false
	}

	if operation != "List" && userFetch.Type != typed {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Can't not %s non %s to property using this endpoint", operation, typed), struct{}{})
		return nil, nil, false
	}

	return property, userFetch, true
}

func augmentProperty(c *gin.Context, typed, operation string, f func(map[string]string, string)) {
	property, userFetch, ok := validateProperty(c, typed, operation)
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

	users, err := models.FetchUserByCriterionMultiple(field, values)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	updateProperty(property)

	models.NewResponse(c, http.StatusOK, fmt.Errorf("%s %s from this property", typed, operation), users)

}

func mapKeysToArray(m map[string]string) []string {
	results := []string{}
	for key := range m {
		results = append(results, key)
	}
	return results
}

func fetchList(c *gin.Context, typed string) {
	property, _, ok := validateProperty(c, typed, "List")
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

	users, err := models.FetchUserByCriterionMultiple(field, values)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("List of %s added to this property", typed), users)
}
