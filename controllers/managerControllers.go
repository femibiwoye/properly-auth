package controllers

import (
	"fmt"
	struct2map "github.com/haibeey/struct2Map"
	"net/http"
	"os"
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

	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}

	property, _ := models.FetchPropertyByCriterion("id", data.ID)
	if property == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Property not found"), errorResponse)
		return
	}
	data.ID = ""
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
	property.ID = data.ID
	err = updateProperty(property)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, property)
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("Property have updated"), response)

}

// RemoveAttachment godoc
// @Summary endpoint to remove an image or document attached to a property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.CreateProperty true "details"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /manager/remove/attachment/ [delete]
// @Security ApiKeyAuth
func RemoveAttachment(c *gin.Context) {
	_, _, ok := checkUser(c, true)
	if !ok {
		return
	}

	data := models.RemoveAttachmentModel{}
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

	property, _ := models.FetchPropertyByCriterion("id", data.PropertyID)
	if property == nil {
		_, ok := errorResponse["PropertyID"]
		if !ok {
			errorResponse["propertyid"] = []string{"Property id doesn't match  any property"}
		}
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Property not found"), errorResponse)
		return
	}

	updated := false
	if strings.Trim(data.AttachmentType, " ") == "documents" {
		for i, doc := range property.Documents {
			if doc == data.AttachmentName {
				property.Documents = utils.RemoveFromArray(property.Documents, i)
				os.RemoveAll(fmt.Sprintf("%spublic/media/%s", os.Getenv("ROOTDIR"), data.AttachmentName))
				updated = true
			}
		}

	} else if strings.Trim(data.AttachmentType, " ") == "images" {
		for i, img := range property.Images {
			if img == data.AttachmentName {
				property.Images = utils.RemoveFromArray(property.Images, i)
				os.RemoveAll(fmt.Sprintf("%spublic/media/%s", os.Getenv("ROOTDIR"), data.AttachmentName))
				updated = true
			}
		}
	} else {
		_, ok := errorResponse["AttachmentType"]
		if !ok {
			errorResponse["attachmenttype"] = []string{fmt.Sprintf("Attachment type not provided or invalid name %s", data.AttachmentType)}
		}

		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Invalid attachment type"), errorResponse)
		return
	}

	err = updateProperty(property)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, property)
		return
	}

	if updated{
		models.NewResponse(c, http.StatusOK, fmt.Errorf("Property is updated"), updated)
	}else{
		models.NewResponse(c, http.StatusOK, fmt.Errorf("Nothing was updated"), updated)
	}

}
