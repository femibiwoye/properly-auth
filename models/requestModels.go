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

type UpdateUserModel struct {
	FirstName   string
	LastName    string
	Dob         string
	PhoneNumber string
	Address     string
}

type ProfileImage struct {
	Image []byte
}

type CreateProperty struct {
	Name    string
	Type    string
	Address string
}

type UpdatePropertyModel struct {
	Name    string
	Type    string
	Address string
	ID      string
}

type AugmentProperty struct {
	UserID     string
	PropertyID string
}

func (a *AugmentProperty) GetUserID() string {
	return a.UserID
}
func (a *AugmentProperty) GetPropertyID() string {
	return a.PropertyID
}

type AddLandLordProperty struct {
	PropertyID   string
	Name         string
	Email        string
	Phone        string
	BusinessName string
}

func (a *AddLandLordProperty) GetUserID() string {
	return ""
}
func (a *AddLandLordProperty) GetPropertyID() string {
	return a.PropertyID
}

type AddTenantProperty struct {
	PropertyID    string
	Name          string
	RentStartDate string
	RentEndDate   string
	Email         string
	Phone         string
}

func (a *AddTenantProperty) GetUserID() string {
	return ""
}
func (a *AddTenantProperty) GetPropertyID() string {
	return a.PropertyID
}

type ListType struct {
	PropertyID string
}

type RemoveAttachmentModel struct {
	PropertyID     string
	AttachmentName string
	AttachmentType string
}

type ScheduleInspectionModel struct {
	PropertyID     string
	AttachmentName string
	AttachmentType string
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
