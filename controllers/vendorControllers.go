package controllers

import (
	"github.com/gin-gonic/gin"
	"properlyauth/models"
)

// AddVendorToProperty godoc
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
// @Router /vendor/property/add/ [put]
// @Security ApiKeyAuth
func AddVendorToProperty(c *gin.Context) {
	augmentProperty(c, models.Tenant, "add", func(m map[string]string, id string) {
		m[id] = id
	})
}

// RemoveVendorFromProperty godoc
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
// @Router /vendor/property/remove/ [put]
// @Security ApiKeyAuth
func RemoveVendorFromProperty(c *gin.Context) {
	augmentProperty(c, models.Tenant, "remove", func(m map[string]string, id string) {
		delete(m, id)
	})
}
