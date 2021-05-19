package general

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"properlyauth/controllers"
	"properlyauth/models"
	"time"

	struct2map "github.com/haibeey/struct2Map"
	"properlyauth/utils"

	"github.com/mitchellh/mapstructure"
)

// MakeComplaints godoc
// @Summary endpoint to add a landloard to a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.ComplaintsModel true "requestdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /make/complaints/ [post]
// @Security ApiKeyAuth
func MakeComplaints(c *gin.Context) {
	data := models.ComplaintsModel{}
	_, _, user, ok := controllers.ValidateProperty(c, &data, false, false, "make", "Complaints")
	if user.Type == models.Manager {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Manager can't make compalints"), struct{}{})
		return
	}
	if !ok {
		return
	}

	complaints := models.Complaints{}
	complaints.CreatedAt = time.Now().Unix()
	complaints.Text = data.Text
	complaints.PropertyId = data.PropertyID
	complaints.CreatedBy = user.ID
	complaints.Status = models.Pending

	if err := models.Insert(&complaints, models.ComplaintsCollectionName); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusCreated, fmt.Errorf("New complaints Activity Created"), complaints)

}

// UpdateComplaints godoc
// @Summary endpoint to remove a landloard from a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.UpdateComplaintsModel true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /update/complaint/ [put]
// @Security ApiKeyAuth
func UpdateComplaints(c *gin.Context) {
	_, _, ok := controllers.CheckUser(c, true)
	if !ok {
		return
	}

	data := models.UpdateComplaintsModel{}
	c.ShouldBindJSON(&data)

	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}

	complaintsM, err := models.FetchDocByCriterion("id", data.ComplaintsID, models.ComplaintsCollectionName)
	if complaintsM == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Complaints not found"), "Invalid Complaints ID")
		return
	}
	complaints, err := models.ToComplaintsFromM(complaintsM)
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

	err = mapstructure.Decode(mapToUpdate, complaints)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	if err := models.UpdateData(complaints, models.ComplaintsCollectionName); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("Complaints Updated"), response)
}

// ListComplaints godoc
// @Summary endpoint to list all the complaints made on a property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.ListType true "requestdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /list/complaint/ [get]
// @Security ApiKeyAuth
func ListComplaints(c *gin.Context) {
	data := models.ListType{}
	_, _, ok := controllers.CheckUser(c, true)
	if !ok {
		return
	}

	c.ShouldBindJSON(&data)

	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}

	if len(errorResponse) > 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided invalid fetch details"), errorResponse)
		return
	}

	response := make(map[string]interface{})
	complaints, err := models.FetchDocByCriterionMultipleAnd(
		[]string{"propertyid", "status"},
		[]string{data.PropertyID, models.Pending},
		models.ComplaintsCollectionName)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}
	response[models.Pending] = complaints
	complaints, err = models.FetchDocByCriterionMultipleAnd(
		[]string{"propertyid", "status"},
		[]string{data.PropertyID, models.Acknowledged},
		models.ComplaintsCollectionName)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}
	response[models.Acknowledged] = complaints
	complaints, err = models.FetchDocByCriterionMultipleAnd(
		[]string{"propertyid", "status"},
		[]string{data.PropertyID, models.Resolved},
		models.ComplaintsCollectionName)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}
	response[models.Resolved] = complaints

	models.NewResponse(c, http.StatusOK, fmt.Errorf("List of complaints no this property"), response)
}

// SaveFiles godoc
// @Summary endpoint to upload albritary files
// @Description
// @Accept  multipart/form-data
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /v1/save/file/ [post]
// @Security ApiKeyAuth
func SaveFiles(c *gin.Context) {
	_, err := utils.DecodeJWT(c)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, struct{}{})
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, struct{}{})
		return
	}
	files, err := controllers.HandleMediaUploads(c, "files", []string{}, form)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}
	models.NewResponse(c, http.StatusOK, fmt.Errorf("Files uploaded"), files)
}

// MakeComplaintsReply godoc
// @Summary endpoint used to add a reply to a complaint
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.ComplaintsReplyModel true "requestdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /v1/make/complaints/reply/ [post]
// @Security ApiKeyAuth
func MakeComplaintsReply(c *gin.Context) {
	res, err := utils.DecodeJWT(c)
	if err != nil {
		models.NewResponse(c, http.StatusUnauthorized, err, struct{}{})
		return
	}

	data := models.ComplaintsReplyModel{}
	c.ShouldBindJSON(&data)
	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}
	if len(errorResponse) > 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided invalid fetch details"), errorResponse)
		return
	}
	_, err = models.GetComplaints("id", data.ComplaintID)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid complaint"), err.Error())
	}

	complaintsReply := models.ComplaintsReply{}
	complaintsReply.CreatedAt = time.Now().Unix()
	complaintsReply.Text = data.Text
	complaintsReply.ComplaintId = data.ComplaintID
	complaintsReply.CreatedBy = res["user_id"]

	if err := models.Insert(&complaintsReply, models.ComplaintsReplyCollectionName); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusCreated, fmt.Errorf("New reply to complaint have been added "), complaintsReply)
}

// ListComplaintsReply godoc
// @Summary endpoint to list all the reply to complaints made on a property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.ListComplaintsReply true "requestdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /v1/list/complaint-reply/ [post]
// @Security ApiKeyAuth
func ListComplaintsReply(c *gin.Context) {
	data := models.ListComplaintsReply{}
	_, _, ok := controllers.CheckUser(c, true)
	if !ok {
		return
	}

	c.ShouldBindJSON(&data)

	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}

	if len(errorResponse) > 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided invalid fetch details"), errorResponse)
		return
	}

	complaints, err := models.FetchDocByCriterionMultiple("complaintid", models.ComplaintsReplyCollectionName, []string{data.ComplaintID})
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("List of reply to complaints "), complaints)
}
