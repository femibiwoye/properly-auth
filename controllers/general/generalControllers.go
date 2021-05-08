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

	if err := controllers.UpdateData(complaints, models.ComplaintsCollectionName); err != nil {
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

	complaints, err := models.FetchDocByCriterionMultiple("propertyid", models.ComplaintsCollectionName, []string{data.PropertyID})
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("List of complaints no this property"), complaints)

}
