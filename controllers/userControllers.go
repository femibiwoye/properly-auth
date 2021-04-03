package controllers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"properlyauth/models"
	"properlyauth/utils"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	struct2map "github.com/haibeey/struct2Map"
	"github.com/mitchellh/mapstructure"
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
	_, isError := ErrorReponses(c, &data, "signup")
	if isError {
		return
	}
	if data.Password != data.ConfirmPassword {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Passwords does not match"), struct{}{})
		return
	}
	user := &models.User{}
	switch strings.ToLower(strings.Trim(data.Type, " ")) {
	case "manager":
		user.Type = models.Manager
	case "landlord":
		user.Type = models.Landlord
	case "tenant":
		user.Type = models.Tenant
	case "vendor":
		user.Type = models.Vendor
	default:
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You have inputed invalid user type"), struct{}{})
		return

	}
	if err := checkmail.ValidateFormat(data.Email); err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Not a valid email"), struct{}{})
		return
	}

	userFound, err := models.FetchDocByCriterion("email", data.Email, models.UserCollectionName)
	if err != nil && err != mongo.ErrNoDocuments {
		models.NewResponse(c, http.StatusInternalServerError, err, "omo")
		return
	}
	if userFound != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Email taken"), struct{}{})
		return
	}

	user.Email = data.Email
	user.FirstName = data.FirstName
	user.LastName = data.LastName
	user.Password = utils.SHA256Hash(data.Password)
	user.CreatedAt = time.Now().Unix()
	user.PUMCCode = utils.GeneratePUMCCode(6)
	if err := models.Insert(user, models.UserCollectionName); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Something went wrong while inserting user"), struct{}{})
		return
	}

	token, err := utils.CreateToken(user.ID)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error creating token"), struct{}{})
		return
	}
	v, err := struct2map.Struct2Map(user)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}
	delete(v, "Password")
	v["token"] = token
	models.NewResponse(c, http.StatusCreated, fmt.Errorf("New User Created"), v)
}

// ResetPassword godoc
// @Summary ResetPassword send link/token to user depending on the platform
// @Description user to reset link or tokne to user mail
// @Tags accounts
// @Accept  json
// @Produce  json
// @Param userDetails body  body models.ResetPassword true "useraccountdetails"
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /reset/update-password/ [put]
// @Security ApiKeyAuth
func ResetPassword(c *gin.Context) {
	data := models.ResetPassword{}
	platform, isError := ErrorReponses(c, &data, "Reset Password")
	if isError {
		return
	}
	userFound, _ := models.FetchDocByCriterion("email", data.Email, models.UserCollectionName)

	if userFound == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("User not found"), nil)
		return
	}

	body := ``
	token := ""

	if platform == "mobile" {
		token = utils.GenerateRandomDigit(6)
		body = fmt.Sprintf(`
			<h1>Reset Password request</h1>
			<p>Your password reset code is %s</p>
		`, token)
		if err := models.SaveToken(data.Email, token, platform); err != nil {
			models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error generating token"), nil)
			return
		}
	} else {
		token = utils.GenerateRandomDigit(15)
		token = base64.StdEncoding.EncodeToString([]byte(token))
		body = fmt.Sprintf(`
		<h1>Reset Password request</h1>
		<a href="%s">Password Reset Link</a>
		`, fmt.Sprintf("http://%s/reset/password/?token=%s&&platform=web", os.Getenv("HOST"), token))
		if err := models.SaveToken(data.Email, token, platform); err != nil {
			models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error generating token"), nil)
			return
		}
	}

	if err := utils.SendMail(data.Email, "Password Reset from Properly", body); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("Reset email sent"), struct{ Token string }{Token: token})
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
// @Router /user/change-password/ [put]
// @Security ApiKeyAuth
func ChangePasswordAuth(c *gin.Context) {
	data := models.ChangeUserPassword{}
	_, isError := ErrorReponses(c, &data, "Change password")
	if isError {
		return
	}

	res, err := utils.DecodeJWT(c)
	if err != nil {
		models.NewResponse(c, http.StatusUnauthorized, err, false)
		return
	}

	userM, _ := models.FetchDocByCriterion("id", res["user_id"], models.UserCollectionName)

	if userM == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("User not found"), false)
		return
	}

	userFetch, err := models.ToUserFromM(userM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}
	if userFetch.Password != utils.SHA256Hash(data.OldPassword) {
		models.NewResponse(c, http.StatusUnauthorized, fmt.Errorf("Wrong old password"), nil)
		return
	}

	userFetch.Password = utils.SHA256Hash(data.Password)

	err = UpdateData(userFetch, models.UserCollectionName)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}
	models.NewResponse(c, http.StatusOK, fmt.Errorf("Password changed"), true)
}

// ChangePasswordFromToken godoc
// @Summary ChangePasswordFromToken changes user password from token sent along
// @Description user to change user password via mail
// @Tags accounts
// @Accept  json
// @Produce  json
// @Param userDetails body models.ChangeUserPasswordFromToken true "userdetails"
// @Success 201 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /reset/validate-token/ [post]
// @Security ApiKeyAuth
func ChangePasswordFromToken(c *gin.Context) {
	var email string
	var password string

	data := models.ChangeUserPasswordFromToken{}
	_, isError := ErrorReponses(c, &data, "Update password")
	if isError {
		return
	}

	tokenData, err := models.FetchToken(data.Email)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	if time.Now().Unix()-tokenData["time"].(int64) > 1800 {
		models.NewResponse(c, http.StatusUnauthorized, fmt.Errorf("Token time is expired"), nil)
		return
	}
	token, ok := tokenData["value"]

	if !ok || token != data.Token {
		models.NewResponse(c, http.StatusUnauthorized, fmt.Errorf("Invalid Token"), nil)
		return
	}
	email = data.Email
	password = data.Password

	userM, _ := models.FetchDocByCriterion("email", email, models.UserCollectionName)
	if userM == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("User not found"), nil)
		return
	}

	userFetch, err := models.ToUserFromM(userM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}
	userFetch.Password = utils.SHA256Hash(password)
	err = UpdateData(userFetch, models.UserCollectionName)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
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
// @Param userDetails body models.LoginData true "useraccountdetails"
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /login/ [post]
// @Security ApiKeyAuth
func SignIn(c *gin.Context) {
	data := models.LoginData{}
	_, isError := ErrorReponses(c, &data, "Login")
	if isError {
		return
	}

	if err := checkmail.ValidateFormat(data.Email); err != nil {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Not a valid email"), struct{ Email []string }{Email: []string{"Invalid Email"}})
		return
	}

	userFound, err := models.FetchDocByCriterion("email", data.Email, models.UserCollectionName)

	if err != nil && err != mongo.ErrNoDocuments {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}

	if userFound == nil {
		models.NewResponse(c, http.StatusUnauthorized, fmt.Errorf("Invalid login details"), nil)
		return
	}

	if value, ok := userFound["password"]; ok && value != utils.SHA256Hash(data.Password) {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Incorrect  Login details"), nil)
		return
	}

	id, _ := userFound["id"]
	token, err := utils.CreateToken(id.(string))
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, fmt.Errorf("Error creating token"), nil)
		return
	}
	v, err := struct2map.Struct2Map(userFound)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}
	delete(v, "Password")
	v["token"] = token
	models.NewResponse(c, http.StatusOK, fmt.Errorf("User signed in"), v)
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
// @Router /user/ [get]
// @Security ApiKeyAuth
func UserProfile(c *gin.Context) {
	_, err := GetPlatform(c)
	if err != nil {
		return
	}
	res, err := utils.DecodeJWT(c)
	if err != nil {
		models.NewResponse(c, http.StatusUnauthorized, err, nil)
		return
	}
	userFetch, _ := models.FetchDocByCriterion("id", res["user_id"], models.UserCollectionName)

	if userFetch == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("user not found"), nil)
		return
	}

	v, err := struct2map.Struct2Map(userFetch)

	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}
	delete(v, "Password")

	models.NewResponse(c, http.StatusOK, fmt.Errorf("User profile"), v)
}

// UpdateProfile godoc
// @Summary endpoint to update user profile
// @Description
// @Tags accounts
// @Accept  json
// @Param  userDetails body models.UpdateUserModel true "useraccountdetails"
// @Produce  json
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /user/update/ [put]
// @Security ApiKeyAuth
func UpdateProfile(c *gin.Context) {
	_, err := GetPlatform(c)
	if err != nil {
		return
	}
	res, err := utils.DecodeJWT(c)
	if err != nil {
		models.NewResponse(c, http.StatusUnauthorized, err, nil)
		return
	}
	data := models.UpdateUserModel{}
	c.ShouldBindJSON(&data)
	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}

	userM, _ := models.FetchDocByCriterion("id", res["user_id"], models.UserCollectionName)

	if userM == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("user not found"), nil)
		return
	}
	userFetch, err := models.ToUserFromM(userM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}
	v, err := struct2map.Struct2Map(&data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
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

	mapstructure.Decode(mapToUpdate, userFetch)
	err = UpdateData(userFetch, models.UserCollectionName)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}

	if len(response) <= 0 {
		models.NewResponse(c, http.StatusOK, fmt.Errorf("Nothing was updated"), response)
	} else {
		models.NewResponse(c, http.StatusOK, fmt.Errorf("User profile update"), response)
	}

}

// UpdateProfileImage godoc
// @Summary endpoint to update user profile
// @Description
// @Tags accounts
// @Accept  multipart/form-data;
// @Produce  json
// @Param  userDetails body models.ProfileImage true "useraccountdetails"
// @Success 200 {object} models.HTTPRes
// @Failure 400 {object} models.HTTPRes
// @Failure 404 {object} models.HTTPRes
// @Failure 500 {object} models.HTTPRes
// @Router /user/update-profile-image/ [put]
// @Security ApiKeyAuth
func UpdateProfileImage(c *gin.Context) {
	_, err := GetPlatform(c)
	if err != nil {
		return
	}
	res, err := utils.DecodeJWT(c)
	if err != nil {
		models.NewResponse(c, http.StatusUnauthorized, err, nil)
		return
	}
	userM, _ := models.FetchDocByCriterion("id", res["user_id"], models.UserCollectionName)

	if userM == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("user not found"), nil)
		return
	}
	userFetch, err := models.ToUserFromM(userM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}

	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		models.NewResponse(c, http.StatusBadRequest, err, struct{ Image []string }{Image: []string{"image file error"}})
		return
	}
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return
	}
	rootDir := os.Getenv("ROOTDIR")
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Base(fileHeader.Filename))
	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/png" {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("The provided file format is not allowed. Please upload a JPEG or PNG image"), struct{}{})
		return
	}
	err = c.SaveUploadedFile(fileHeader, fmt.Sprintf("%s/public/media/%s", rootDir, filename))
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}
	userFetch.ProfileImageURL = filename

	err = UpdateData(userFetch, models.UserCollectionName)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return
	}
	models.NewResponse(c, http.StatusOK, fmt.Errorf("Profile image updated"), true)
}
