package manager

import (
	"github.com/gin-gonic/gin"
	"properlyauth/controllers"
	"properlyauth/models"
)

// AddTenantToProperty godoc
// @Summary endpoint to add a tenant to a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.AugmentProperty true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /tenant/property/add/ [post]
// @Security ApiKeyAuth
func AddTenantToProperty(c *gin.Context) {
	controllers.AugmentProperty(c, models.Tenant, "add", func(m map[string]string, id string) {
		m[id] = id
	})
}

// RemoveTenantFromProperty godoc
// @Summary endpoint to remove a tanent from a property. Only manager are capable of adding landlord property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.AugmentProperty true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /tenant/property/remove/ [delete]
// @Security ApiKeyAuth
func RemoveTenantFromProperty(c *gin.Context) {
	controllers.AugmentProperty(c, models.Tenant, "remove", func(m map[string]string, id string) {
		delete(m, id)
	})
}

// ListTenantFromProperty godoc
// @Summary list all tenant in a property
// @Description
// @Tags accounts
// @Accept  json
// @Param  details body models.ListType true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /tenant/property/list/ [get]
// @Security ApiKeyAuth
func ListTenantFromProperty(c *gin.Context) {
	controllers.FetchList(c, models.Tenant)
}
