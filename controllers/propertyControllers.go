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
	struct2map "github.com/haibeey/struct2Map"
	"github.com/mitchellh/mapstructure"
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

func checkUser(c *gin.Context) (*models.User, string, bool) {
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
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("user not found"), struct{}{})
		return nil, "", false
	}

	if userFetch.Type != models.Manager {
		models.NewResponse(c, http.StatusUnauthorized, fmt.Errorf("Only managers can create and change properties"), userFetch)
		return nil, "", false
	}

	return userFetch, platform, true
}

func augmentProperty(c *gin.Context, typed, operation string, f func(map[string]string, string)) {
	_, _, ok := checkUser(c)
	if !ok {
		return
	}

	data := models.AddLandlord{}
	c.ShouldBindJSON(&data)

	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}

	if len(errorResponse) > 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided invalid %s property details", operation), errorResponse)
		return
	}

	property, _ := models.FetchPropertyByCriterion("id", data.PropertyID)
	if property == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Property not found"), data)
		return
	}

	userFetch, _ := models.FetchUserByCriterion("id", data.UserID)

	if userFetch == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("user to %s not found", operation), struct{}{})
		return
	}

	if userFetch.Type != typed {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Can't not %s non %s to property using this endpoint", operation, typed), struct{}{})
		return
	}

	if typed == models.Landlord {
		f(property.Landlord, userFetch.ID)
	} else if typed == models.Tenant {
		f(property.Tenants, userFetch.ID)
	}

	updateProperty(property)

	models.NewResponse(c, http.StatusOK, fmt.Errorf("New %s added to this property", typed), struct{}{})

}

// CreateProperty godoc
// @Summary endpoint to Create a property. Only manager are capable of creating property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.CreateProperty true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /create/property/ [put]
// @Security ApiKeyAuth
func CreateProperty(c *gin.Context) {
	userFetch, _, ok := checkUser(c)
	if !ok {
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, struct{}{})
		return
	}

	data := models.CreateProperty{
		Name:    strings.Join(form.Value["name"], "\n"),
		Type:    strings.Join(form.Value["type"], "\n"),
		Address: strings.Join(form.Value["address"], "\n"),
	}
	_, isError := errorReponses(c, &data, "Create Property")
	if isError {
		return
	}

	images, err := handleMediaUploads(c, "images", form)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	if len(images)<=0{
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("No image provided"), struct{}{})
		return
	}

	documents, err := handleMediaUploads(c, "documents", form)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	property := models.Property{}
	property.Documents = documents
	property.Images = images
	property.Address = data.Address
	property.Name = data.Name
	property.Type = data.Type
	property.Landlord = make(map[string]string)
	property.Tenants = make(map[string]string)
	property.CreatedAt = time.Now().Unix()
	property.CreatedBy = userFetch.ID
	property.Status = "created"

	if err := models.InsertProperty(&property); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusCreated, fmt.Errorf("New Property Created"), property)
}

// UpdatePropertyRoute godoc
// @Summary endpoint to edit a property field. Only manager are capable of updating property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.CreateProperty true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /update/property/ [put]
// @Security ApiKeyAuth
func UpdatePropertyRoute(c *gin.Context) {
	_, _, ok := checkUser(c)
	if !ok {
		return
	}

	data := models.UpdatePropertyModel{}
	c.ShouldBindJSON(&data)

	property, _ := models.FetchPropertyByCriterion("id", data.ID)
	if property == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Property not found"), data)
		return
	}
	data.ID = ""
	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}

	v, err := struct2map.Struct2Map(&data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}

	mapToUpdate := make(map[string]interface{})
	response := make(map[string]interface{})
	for key, value := range v {
		_, ok := errorResponse[key]
		if !ok {
			mapToUpdate[key] = value
			response[key] = []string{fmt.Sprintf("%s has been updated to %s", key, value)}
		}
	}

	if len(response) <= 0 {
		models.NewResponse(c, http.StatusOK, fmt.Errorf("Nothing was updated"), response)
		return
	}
	data.ID = property.ID
	mapstructure.Decode(mapToUpdate, property)
	err = updateProperty(property)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, response)
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("User profile update"), response)

}

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
// @Router /property/add-landlord/ [put]
// @Security ApiKeyAuth
func AddLandlordToProperty(c *gin.Context) {
	augmentProperty(c, models.Landlord, "add", func(m map[string]string, id string) {
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
// @Router /property/remove-landlord/ [put]
// @Security ApiKeyAuth
func RemoveLandlordFromProperty(c *gin.Context) {
	augmentProperty(c, models.Landlord, "remove", func(m map[string]string, id string) {
		delete(m, id)
	})
}

// AddTenantToProperty godoc
// @Summary endpoint to add a tenant to a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.AddLandlord true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /property/add-tenant/ [put]
// @Security ApiKeyAuth
func AddTenantToProperty(c *gin.Context) {
	augmentProperty(c, models.Tenant, "add", func(m map[string]string, id string) {
		m[id] = id
	})
}

// RemoveTenantFromProperty godoc
// @Summary endpoint to remove a tanent from a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.AddLandlord true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /property/remove-tenant/ [put]
// @Security ApiKeyAuth
func RemoveTenantFromProperty(c *gin.Context) {
	augmentProperty(c, models.Tenant, "remove", func(m map[string]string, id string) {
		delete(m, id)
	})
}
