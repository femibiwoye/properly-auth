package controllers

import (
	"fmt"
	struct2map "github.com/haibeey/struct2Map"
	"net/http"
	"properlyauth/models"
	"properlyauth/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mitchellh/mapstructure"
)

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
// @Router /manager/create/property/ [put]
// @Security ApiKeyAuth
func CreateProperty(c *gin.Context) {
	userFetch, _, ok := checkUser(c, true)
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

	if len(images) <= 0 {
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
	property.Landlords = make(map[string]string)
	property.Tenants = make(map[string]string)
	property.Vendors = make(map[string]string)
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
// @Router /manager/update/property/ [put]
// @Security ApiKeyAuth
func UpdatePropertyRoute(c *gin.Context) {
	_, _, ok := checkUser(c, true)
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
