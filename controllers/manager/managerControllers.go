package manager

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
	"properlyauth/controllers"
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
	userFetch, _, ok := controllers.CheckUser(c, true)
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
	_, isError := controllers.ErrorReponses(c, &data, "Create Property")
	if isError {
		return
	}

	images, err := controllers.HandleMediaUploads(c, "images", form)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	if len(images) <= 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("No image provided"), struct{}{})
		return
	}

	documents, err := controllers.HandleMediaUploads(c, "documents", form)
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

	if err := models.Insert(&property, models.PropertyCollectionName); err != nil {
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
	_, _, ok := controllers.CheckUser(c, true)
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

	propertyM, _ := models.FetchDocByCriterion("id", data.ID, models.PropertyCollectionName)
	if propertyM == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Property not found"), errorResponse)
		return
	}
	property, err := models.ToPropertyFromM(propertyM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
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
	err = controllers.UpdateData(property, models.PropertyCollectionName)
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
	_, _, ok := controllers.CheckUser(c, true)
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

	propertyM, _ := models.FetchDocByCriterion("id", data.PropertyID, models.PropertyCollectionName)
	if propertyM == nil {
		_, ok := errorResponse["PropertyID"]
		if !ok {
			errorResponse["propertyid"] = []string{"Property id doesn't match  any property"}
		}
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Property not found"), errorResponse)
		return
	}

	property, err := models.ToPropertyFromM(propertyM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
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

	err = controllers.UpdateData(property, models.PropertyCollectionName)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, property)
		return
	}

	if updated {
		models.NewResponse(c, http.StatusOK, fmt.Errorf("Property is updated"), updated)
	} else {
		models.NewResponse(c, http.StatusOK, fmt.Errorf("Nothing was updated"), updated)
	}

}

// ScheduleInspection godoc
// @Summary endpoint to create an inspection on a property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.InspectionModel true "request details"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /manager/inspection/schedule/ [put]
// @Security ApiKeyAuth
func ScheduleInspection(c *gin.Context) {
	data := models.InspectionModel{}
	_, _, user, ok := controllers.ValidateProperty(c, &data, false, "Schedule", "Inspection")
	if !ok {
		return
	}

	inspection := models.Inspection{}
	inspection.CreatedAt = time.Now().Unix()
	inspection.DueTime = data.Date
	inspection.Text = data.Text
	inspection.PropertyId = data.PropertyID
	inspection.CreatedBy = user.ID

	if err := models.Insert(&inspection, models.InspectionCollectionaName); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusCreated, fmt.Errorf("New Inspection Activity Created"), inspection)
}

// ScheduleInspection godoc
// @Summary endpoint update an existing inspection
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.UpdateInspectionModel true "request details"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /manager/inspection/update/ [post]
// @Security ApiKeyAuth
func UpdateInspection(c *gin.Context) {

	_, _, ok := controllers.CheckUser(c, true)
	if !ok {
		return
	}

	data := models.UpdateInspectionModel{}
	c.ShouldBindJSON(&data)

	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}

	inspectionM, err := models.FetchDocByCriterion("id", data.InspectionID, models.InspectionCollectionaName)
	if inspectionM == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Inspection not found"), errorResponse)
		return
	}
	inspection, err := models.ToInspectionFromM(inspectionM)
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

	mapstructure.Decode(mapToUpdate, inspection)
	if err := controllers.UpdateData(inspection, models.InspectionCollectionaName); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("Inspection Updated"), response)
}

// ScheduleInspection godoc
// @Summary endpoint to remove an inspection
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.InspectionDeleteModel true "details"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /manager/inspection/delete/ [delete]
// @Security ApiKeyAuth
func DeleteInspection(c *gin.Context) {
	_, _, ok := controllers.CheckUser(c, true)
	if !ok {
		return
	}

	data := models.InspectionDeleteModel{}
	c.ShouldBindJSON(&data)

	inspectionM, err := models.FetchDocByCriterion("id", data.InspectionID, models.InspectionCollectionaName)

	if inspectionM == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Inspection not found"), struct{}{})
		return
	}
	inspection, err := models.ToInspectionFromM(inspectionM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	if err := models.Delete(inspection, models.InspectionCollectionaName); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("Inspection Deleted"), struct{}{})
}
