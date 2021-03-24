package models

import (
	"github.com/gin-gonic/gin"
)

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpData struct {
	Type            string `json:"type"`
	FirstName       string `json:"firstname"`
	LastName        string `json:"lastname"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmpassword"`
}

type ResetPassword struct {
	Email string `json:"email"`
}
type TokenAndPhoneData struct {
	Phone string `json:"phone"`
	Token string `json:"token"`
}

type ChangeUserPassword struct {
	Email       string `json:"email"`
	OldPassword string `json:"oldpassword"`
	Password    string `json:"password"`
}

type ChangeUserPasswordFromToken struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Password string `json:"password"`
}

type SignupResponse struct {
	Success string
}

type EnterTokenResponse struct {
	Success string
	Token   string
}

type CompleteSignUp struct {
	Fullname string
	Email    string
	Region   string
}

// NewResponse example
func NewResponse(ctx *gin.Context, status int, err error, data interface{}) {
	er := HTTPRes{
		Code:    status,
		Message: err.Error(),
		Data:    data,
	}
	ctx.JSON(status, er)
}

// HTTPRes example
type HTTPRes struct {
	Code    int         `json:"code" example:""`
	Message string      `json:"message" example:"status bad request"`
	Data    interface{} `json:"data"`
}
