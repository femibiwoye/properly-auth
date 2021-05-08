package manager

import (
	"fmt"
	struct2map "github.com/haibeey/struct2Map"
	"net/http"
	"properlyauth/controllers"
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

	images, err := controllers.HandleMediaUploads(c, "images", []string{"image/jpeg", "image/png"}, form)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Image type not accepted"), err.Error())
		return
	}

	if len(images) <= 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("No image provided"), struct{}{})
		return
	}

	documents, err := controllers.HandleMediaUploads(c, "documents", []string{
		"application/msword", "application/pdf", "application/zip"},
		form)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("document type doesn't match"), err.Error())
		return
	}

	property := models.Property{}
	property.Documents = documents
	property.Images = images
	property.Address = data.Address
	property.Name = data.Name
	property.Type = data.Type
	property.Forms = []string{}
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

	form, err := c.MultipartForm()
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, "brah")
		return
	}
	data := models.UpdatePropertyModel{
		Name:    strings.Join(form.Value["name"], "\n"),
		Type:    strings.Join(form.Value["type"], "\n"),
		Address: strings.Join(form.Value["address"], "\n"),
		ID:      strings.Join(form.Value["id"], "\n"),
	}

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

	images, err := controllers.HandleMediaUploads(c, "images", []string{"image/jpeg", "image/png"}, form)
	documents, err := controllers.HandleMediaUploads(c, "documents", []string{
		"application/msword", "application/pdf", "application/zip"},
		form)

	data.ID = property.ID
	err = mapstructure.Decode(mapToUpdate, property)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error converting property"), err)
		return
	}
	property.ID = data.ID
	if len(images) > 0 {
		property.Images = images
	}
	if len(documents) > 0 {
		property.Documents = documents
	}

	err = controllers.UpdateData(property, models.PropertyCollectionName)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, property)
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("Property have been updated"), response)
}

// RemoveAttachment godoc
// @Summary endpoint to remove an image or document attached to a property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.RemoveAttachmentModel true "details"
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
		removed := 0
		for i, doc := range property.Documents {
			if doc == data.AttachmentName {
				property.Documents = utils.RemoveFromArray(property.Documents, i-removed)
				removed++
				updated = true
			}
		}

	} else if strings.Trim(data.AttachmentType, " ") == "images" {
		removed := 0
		for i, img := range property.Images {
			if img == data.AttachmentName {
				property.Images = utils.RemoveFromArray(property.Images, i-removed)
				removed++
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
// @Router /manager/inspection/schedule/ [post]
// @Security ApiKeyAuth
func ScheduleInspection(c *gin.Context) {
	data := models.InspectionModel{}
	_, _, user, ok := controllers.ValidateProperty(c, &data, false, false, "Schedule", "Inspection")
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
// @Router /manager/inspection/update/ [put]
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

	err = mapstructure.Decode(mapToUpdate, inspection)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	if err := controllers.UpdateData(inspection, models.InspectionCollectionaName); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("Inspection Updated"), inspection)
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

// ListProperties godoc
// @Summary endpoint to list all the property created by user
// @Description
// @Tags accounts
// @Accept  json
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /list/properties/ [get]
// @Security ApiKeyAuth
func ListProperties(c *gin.Context) {
	user, _, ok := controllers.CheckUser(c, true)
	if !ok {
		return
	}

	fmt.Println(user.ID)
	properties, err := models.FetchDocByCriterionMultiple("createdby", models.PropertyCollectionName, []string{user.ID})
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("List of  properties"), properties)

}

// ListInspection godoc
// @Summary endpoint to list all the Inspection created by user
// @Description
// @Tags accounts
// @Accept  json
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /list/inspection/ [get]
// @Security ApiKeyAuth
func ListInspection(c *gin.Context) {
	user, _, ok := controllers.CheckUser(c, true)
	if !ok {
		return
	}

	properties, err := models.FetchDocByCriterionMultiple("createdby", models.InspectionCollectionaName, []string{user.ID})
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("List of  inspection"), properties)

}

// UploadAgreementForm godoc
// @Summary endpoint is used to create a new form for other user
// @Description
// @Tags accounts
// @Accept  json
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /upload/upload/form/ [post]
// @Security ApiKeyAuth
func UploadAgreementForm(c *gin.Context) {
	_, _, ok := controllers.CheckUser(c, true)
	if !ok {
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, struct{}{})
		return
	}
	data := models.ListType{
		PropertyID: strings.Join(form.Value["propertyid"], "\n"),
	}

	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}
	propertyM, _ := models.FetchDocByCriterion("id", data.PropertyID, models.PropertyCollectionName)
	if propertyM == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Property not found"), errorResponse)
		return
	}
	property, err := models.ToPropertyFromM(propertyM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}

	forms, err := controllers.HandleMediaUploads(c,
		"form",
		[]string{"application/msword", "application/pdf", "application/zip"},
		form)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, struct{}{})
		return
	}

	if len(forms) <= 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("no form data uploaded"), struct{}{})
		return
	}

	property.Forms = append(property.Forms, forms...)
	err = controllers.UpdateData(property, models.PropertyCollectionName)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, property)
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("Form Uploaded"), forms)

}
