package controllers

import (
	"encoding/base64"
	"fmt"
	"github.com/badoux/checkmail"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"os"
	"properlyauth/models"
	"properlyauth/utils"
	"strings"
	"time"

	"github.com/haibeey/struct2Map"
)

// SignUp godoc
// @Summary is the endpoint for user signup.
// A user send a his/her phone number to this endpoint to receive token
// @Description SignUp user with email or name
// @Accept  json
// @Produce  json
// @Param  userDetails body models.SignUpData true "useraccountdetails"
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /signup/ [post]
func SignUp(c *gin.Context) {
	data := models.SignUpData{}

	err := c.ShouldBindJSON(&data)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Data not sent %s", err), nil)
		return
	}

	if len(data.Password) <= 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("No password passed"), nil)
		return
	}

	if data.Password != data.ConfirmPassword {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Password does not match"), nil)
		return
	}

	userEmail := true
	if err := checkmail.ValidateFormat(data.Email); err != nil {
		if len(data.Name) <= 0 {
			models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Not a valid email or name"), nil)
			return
		}
		userEmail = false
	}

	user := &models.User{}
	var userFound interface{}
	if userEmail {
		userFound, err = models.FetchUserByCriterion("email", data.Email)
	} else {
		userFound, err = models.FetchUserByCriterion("name", data.Name)
	}

	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}
	if userFound != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Email  or name taken"), nil)
		return
	}
	user.Email = data.Email
	user.Name = data.Name

	user.Password = utils.SHA256Hash(user.Password)
	user.CreatedAt = time.Now().Unix()
	if err := models.InsertUser(user); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Something went wrong while inserting user"), nil)
		return
	}

	token, err := utils.CreateToken(user.ID)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error creating token"), nil)
		return
	}
	tknData := struct {
		Token string
	}{
		Token: token,
	}
	models.NewResponse(c, http.StatusCreated, fmt.Errorf("New User Created "), tknData)
}

// ResetPassword godoc
// @Summary ResetPassword send link/token to user depending on the platform
// @Description user to reset link or tokne to user mail
// @Tags accounts
// @Accept  json
// @Produce  json
// @Param userDetails body  body models.ChangeUserPassword true "useraccountdetails"
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /reset/password/ [post]
// @Security ApiKeyAuth
func ResetPassword(c *gin.Context) {
	data := models.ResetPassword{}
	err := c.ShouldBindJSON(&data)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid data sent"), nil)
		return
	}

	query := c.Request.URL.Query()
	platform, ok := query["platform"]

	if !ok || len(platform) <= 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("No query sent for platform type sent"), nil)
		return
	}

	userFound, err := models.FetchUserByCriterion("email", data.Email)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}

	if userFound != nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("User not found"), nil)
		return
	}

	body := ``
	token := utils.GenerateRandomDigit(6)
	if strings.Trim(platform[0], " ") == "mobile" {
		body = fmt.Sprintf(`
			<h1>Reset Password request</h1>
			<p>Your password reset code is %s</p>
		`, token)
		if err := models.SaveToken(data.Email, token, strings.Trim(platform[0], " ")); err != nil {
			models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error generating token"), nil)
			return
		}
	} else {
		tokenHash := base64.StdEncoding.EncodeToString([]byte(token))
		body = fmt.Sprintf(`
		<h1>Reset Password request</h1>
		<a href="%s">Password Reset Link</a>
		`, fmt.Sprintf("%s/reset/password/?token=%s&&platform=web", os.Getenv("HOST"), tokenHash))
		if err := models.SaveToken(data.Email, tokenHash, strings.Trim(platform[0], " ")); err != nil {
			models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error generating token"), nil)
			return
		}
	}

	if err := utils.SendMail(data.Email, "Password Reset from Properly", body); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error sending email"), nil)
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("Reset email sent"), nil)
	return
}

// ChangePasswordAuth godoc
// @Summary ChangePasswordAuth changes a user password for an authorized user
// @Description user to change user password via mail
// @Tags accounts
// @Accept  json
// @Produce  json
// @Param userDetails body models.ChangeUserPassword true "userdetails"
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /change/password/auth/ [post]
// @Security ApiKeyAuth
func ChangePasswordAuth(c *gin.Context) {
	res, err := utils.DecodeJWT(c)

	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	data := models.ChangeUserPassword{}
	err = c.ShouldBindJSON(&data)

	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid data sent"), nil)
		return
	}

	userFetch, err := models.FetchUserByCriterion("id", res["user_id"])

	if err != nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("User not found"), nil)
		return
	}

	if userFetch.Password != utils.SHA256Hash(data.OldPassword) {
		models.NewResponse(c, http.StatusUnauthorized, fmt.Errorf("Password does not match"), nil)
		return
	}

	userFetch.Password = utils.SHA256Hash(data.Password)
	uB, err := bson.Marshal(userFetch)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid data sent can't parse data"), nil)
		return
	}
	var update bson.M
	err = bson.Unmarshal(uB, &update)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, nil)
		return
	}
	err = models.UpdateUser(userFetch, bson.D{{Key: "$set", Value: update}})
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, nil)
		return
	}
	models.NewResponse(c, http.StatusOK, fmt.Errorf("Password changed"), nil)
}

// ChangePasswordFromToken godoc
// @Summary ChangePasswordFromToken changes user password from token sent along
// @Description user to change user password via mail
// @Tags accounts
// @Accept  json
// @Produce  json
// @Param userDetails body models.ChangeUserPasswordFromToken true "userdetails"
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /change/password/token/ [post]
// @Security ApiKeyAuth
func ChangePasswordFromToken(c *gin.Context) {

	query := c.Request.URL.Query()
	platform, ok := query["platform"]

	if !ok || len(platform) <= 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("No query sent for platform type sent"), nil)
		return
	}

	var email string
	var password string

	data := models.ChangeUserPasswordFromToken{}
	err := c.ShouldBindJSON(&data)

	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid data sent"), nil)
		return
	}

	tokenData, err := models.FetchToken(data.Email)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}
	token, ok := tokenData["value"]
	if !ok || token != data.Token {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid Token"), nil)
		return
	}
	email = data.Email
	password = data.Password

	userFetch, err := models.FetchUserByCriterion("email", email)
	if err != nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("User not found"), nil)
		return
	}

	userFetch.Password = utils.SHA256Hash(password)
	uB, err := bson.Marshal(userFetch)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid data sent can't parse data"), nil)
		return
	}
	var update bson.M
	err = bson.Unmarshal(uB, &update)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, nil)
		return
	}
	err = models.UpdateUser(userFetch, bson.D{{Key: "$set", Value: update}})
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, nil)
		return
	}
	models.TakeOutToken(data.Email)
	models.NewResponse(c, http.StatusOK, fmt.Errorf("Password changed"), nil)

}

// SignIn godoc
// @Summary SignIn is used to login a user
// @Description login a user
// @Tags accounts
// @Accept  json
// @Produce  json
// @Param userDetails body models.SignUpData true "useraccountdetails"
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /signin/ [post]
// @Security ApiKeyAuth
func SignIn(c *gin.Context) {
	data := models.LoginData{}

	err := c.ShouldBindJSON(&data)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Data not sent %s", err), nil)
		return
	}

	if len(data.Email) <= 0 && len(data.Name) <= 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("No email/name passed"), nil)
		return
	}

	if len(data.Password) <= 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("No password passed"), nil)
		return
	}

	userEmail := true
	if err := checkmail.ValidateFormat(data.Email); err != nil {
		if len(data.Name) <= 0 {
			models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Not a valid email or name"), nil)
			return
		}
		userEmail = false
	}

	userFound := &models.User{}
	if userEmail {
		userFound, _ = models.FetchUserByCriterion("email", data.Email)
	} else {
		userFound, _ = models.FetchUserByCriterion("name", data.Name)
	}

	if userFound.Password != utils.SHA256Hash(data.Password) {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Incorrect password"), nil)
		return
	}

	token, err := utils.CreateToken(userFound.ID)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error creating token"), nil)
		return
	}
	tknData := struct {
		Token string
	}{
		Token: token,
	}
	models.NewResponse(c, http.StatusCreated, fmt.Errorf("USer signed in"), tknData)
}

// GeneratePUMC godoc
// @Summary GeneratePUMC generates a unigue code for each user for later user
// @Description pumc code generations
// @Tags accounts
// @Accept  json
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /generate/pumc/ [get]
// @Security ApiKeyAuth
func GeneratePUMC(c *gin.Context) {
	res, err := utils.DecodeJWT(c)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid data sent"), nil)
		return
	}
	userFetch, err := models.FetchUserByCriterion("id", res["user_id"])
	if err != nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("User not found"), nil)
		return
	}

	userFetch.PUMCCode = utils.GeneratePUMCCode(10)
	uB, err := bson.Marshal(userFetch)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid data sent can't parse data"), nil)
		return
	}
	var update bson.M
	err = bson.Unmarshal(uB, &update)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, nil)
		return
	}
	err = models.UpdateUser(userFetch, bson.D{{Key: "$set", Value: update}})
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, nil)
		return
	}

	pumcData := struct {
		PUMCCode string
	}{
		PUMCCode: userFetch.PUMCCode,
	}
	models.NewResponse(c, http.StatusOK, fmt.Errorf("Password changed"), pumcData)
}

// UserProfile godoc
// @Summary returns the user profile
// @Description
// @Tags accounts
// @Accept  json
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /profile/ [get]
// @Security ApiKeyAuth
func UserProfile(c *gin.Context) {
	res, err := utils.DecodeJWT(c)
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid data sent"), nil)
		return
	}
	userFetch, err := models.FetchUserByCriterion("id", res["user_id"])
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}

	if userFetch == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("user not found"), nil)
		return
	}

	v, err := struct2map.Struct2Map(userFetch)

	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}
	delete(v, "password")

	models.NewResponse(c, http.StatusOK, fmt.Errorf("User profile"), v)
}
