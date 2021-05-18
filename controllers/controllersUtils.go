package controllers

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"properlyauth/models"
	"properlyauth/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/haibeey/struct2Map"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type managerRequestData interface {
	GetUserID() string
	GetPropertyID() string
}

//HandleMediaUploads helper function to upload media files
func HandleMediaUploads(c *gin.Context, nameOf string, acceptableDocType []string, form *multipart.Form) ([]string, error) {
	files := form.File[nameOf]
	names := []string{}
	errors := []error{}
	s, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_S3_KEY"),
			os.Getenv("AWS_S3_SECRET"),
			""),
	})
	if err != nil {
		return names, err
	}
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			errors = append(errors, err)
		}
		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			errors = append(errors, err)
		}
		filetype := http.DetectContentType(buff)
		fileTypeGood := false
		for _, docType := range acceptableDocType {
			if filetype == docType {
				fileTypeGood = true
			}
		}
		if !fileTypeGood {
			errors = append(errors, fmt.Errorf("FileType not accepted"))
		}
		filename, err := UploadFileToS3(s, file, fileHeader)
		defer file.Close()
		names = append(names, filename)
	}
	if len(errors) > 0 {
		return names, errors[0]
	}
	return names, nil
}

//HandleMediaUpload helper function to upload media file
func HandleMediaUpload(c *gin.Context, fileHeader *multipart.FileHeader) (string, error) {
	var filename string
	var err error
	s, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_S3_KEY"),
			os.Getenv("AWS_S3_SECRET"),
			""),
	})
	if err != nil {
		return filename, err
	}
	file, err := fileHeader.Open()
	if err != nil {
		return filename, err
	}
	return UploadFileToS3(s, file, fileHeader)
}

//CheckUser helper function to validate a user and optional check if he is manager
func CheckUser(c *gin.Context, checkManager bool) (*models.User, string, bool) {
	platform, err := GetPlatform(c)
	if err != nil {
		return nil, "", false
	}
	res, err := utils.DecodeJWT(c)
	if err != nil {
		models.NewResponse(c, http.StatusUnauthorized, err, struct{}{})
		return nil, "", false
	}

	userM, err := models.FetchDocByCriterion("id", res["user_id"], models.UserCollectionName)

	if userM == nil {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("user not found"), struct{}{})
		return nil, "", false
	}
	userFetch, err := models.ToUserFromM(userM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return nil, "", false
	}
	if checkManager {
		if userFetch.Type != models.Manager {
			models.NewResponse(c, http.StatusUnauthorized, fmt.Errorf("Only managers have access to this resource"), struct{}{})
			return nil, "", false
		}
	}

	return userFetch, platform, true
}

//ValidateProperty validate a give property id and user
func ValidateProperty(c *gin.Context,
	data managerRequestData,
	checkUserID bool, // used to confirm if a user id should be check in the request
	checkManager bool,
	typed, // use to indicate what type of operation we are operating
	operation string) (
	*models.Property,
	*models.User, // The user to change his/her  details
	*models.User, // The user making the request (the manager)
	bool,
) {
	user, _, ok := CheckUser(c, checkManager)
	if !ok {
		return nil, nil, nil, ok
	}
	c.ShouldBindJSON(&data)

	errorResponse, err := utils.MissingDataResponse(data)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, false)
		return nil, nil, nil, false
	}

	// we can omit user ID for some request that don't need it
	if !checkUserID {
		delete(errorResponse, "userid")
	}
	if len(errorResponse) > 0 {
		if operation != "List" {
			models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided invalid %s property details", operation), errorResponse)
			return nil, nil, nil, false
		}
	}

	propertyM, _ := models.FetchDocByCriterion("id", data.GetPropertyID(), models.PropertyCollectionName)
	if propertyM == nil {
		errorResponse["propertyid"] = []string{"Property id doesn't match  any property"}
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Property not found"), errorResponse)
		return nil, nil, nil, false
	}
	property, err := models.ToPropertyFromM(propertyM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return nil, nil, nil, false
	}
	if !checkUserID {
		return property, nil, user, true
	}
	userM, err := models.FetchDocByCriterion("id", data.GetUserID(), models.UserCollectionName)
	if userM == nil {
		errorResponse["userid"] = []string{"User id doesn't match  any user"}
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("The User to %s not found", operation), errorResponse)
		return nil, nil, nil, false
	}
	userFetch, err := models.ToUserFromM(userM)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, nil)
		return nil, nil, nil, false
	}
	if userFetch.Type != typed {
		models.NewResponse(c, http.StatusNotFound, fmt.Errorf("Can't not %s non %s to property using this endpoint", operation, typed), struct{}{})
		return nil, nil, nil, false
	}
	return property, userFetch, user, true
}

//AugmentProperty helper function to change sub data of a property
func AugmentProperty(c *gin.Context, typed, operation string, f func(map[string]string, string)) {
	data := models.AugmentProperty{}
	property, userFetch, _, ok := ValidateProperty(c, &data, true, false, typed, operation)
	if !ok {
		return
	}
	field := "id"
	values := []string{}
	switch typed {
	case models.Landlord:
		f(property.Landlords, userFetch.ID)
		values = mapKeysToArray(property.Landlords)
	case models.Tenant:
		f(property.Tenants, userFetch.ID)
		values = mapKeysToArray(property.Tenants)
	case models.Vendor:
		f(property.Vendors, userFetch.ID)
		values = mapKeysToArray(property.Vendors)
	}

	users, err := models.FetchDocByCriterionMultiple(field, models.UserCollectionName, values)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}
	if err = models.UpdateData(property, models.PropertyCollectionName); err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}
	models.NewResponse(c, http.StatusOK, fmt.Errorf("%s %s from this property", typed, operation), users)
}

func mapKeysToArray(m map[string]string) []string {
	results := []string{}
	for key := range m {
		results = append(results, key)
	}
	return results
}

func FetchList(c *gin.Context, typed string) {
	data := models.AugmentProperty{}
	property, _, _, ok := ValidateProperty(c, &data, false, false, typed, "List")
	if !ok {
		return
	}

	field := "id"
	values := []string{}
	switch typed {
	case models.Landlord:
		values = mapKeysToArray(property.Landlords)
	case models.Tenant:
		values = mapKeysToArray(property.Tenants)
	case models.Vendor:
		values = mapKeysToArray(property.Vendors)
	}

	users, err := models.FetchDocByCriterionMultiple(field, models.UserCollectionName, values)
	if err != nil {
		models.NewResponse(c, http.StatusInternalServerError, err, struct{}{})
		return
	}

	models.NewResponse(c, http.StatusOK, fmt.Errorf("List of %s added to this property", typed), users)
}

//GetPlatform get the platform type for this request
func GetPlatform(c *gin.Context) (string, error) {
	query := c.Request.URL.Query()
	platform, ok := query["platform"]

	if !ok || len(platform) <= 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("Invalid platform"), nil)
		return "", fmt.Errorf("No query sent for platform type sent")
	}
	return strings.Trim(platform[0], " "), nil
}

//ErrorReponses generate error response for empty data
func ErrorReponses(c *gin.Context, data interface{}, api string) (string, bool) {
	platform, err := GetPlatform(c)
	if err != nil {
		return platform, true
	}
	c.ShouldBindJSON(&data)
	errorReponse, err := utils.MissingDataResponse(data)
	if len(errorReponse) > 0 {
		models.NewResponse(c, http.StatusBadRequest, fmt.Errorf("You provided invalid %s details", api), errorReponse)
		return platform, true
	}
	return platform, false
}

// UploadFileToS3 saves a file to aws bucket and returns the url to the file and an error if there's any
func UploadFileToS3(s *session.Session, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	uploader := s3manager.NewUploader(s)
	tempFileName := "users/" + utils.GenerateRandom(10) + filepath.Ext(fileHeader.Filename)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("properlyng"),
		ACL:    aws.String("public-read"),
		Key:    aws.String(tempFileName),
		Body:   file,
	})
	if err != nil {
		return "", err
	}
	return "https://properlyng.s3-eu-west-2.amazonaws.com/" + tempFileName, err
}

type NamedID struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func ConvertPropertyList(p *models.Property) (map[string]interface{}, error) {
	value, err := struct2map.Struct2Map(p)
	if err != nil {
		return nil, err
	}
	landlord := []NamedID{}
	tenants := []NamedID{}
	vendors := []NamedID{}
	managers := []NamedID{}
	for id, name := range p.Landlords {
		landlord = append(landlord, NamedID{name, id})
	}
	for id, name := range p.Managers {
		managers = append(managers, NamedID{name, id})
	}
	for id, name := range p.Vendors {
		vendors = append(vendors, NamedID{name, id})
	}
	for id, name := range p.Tenants {
		tenants = append(tenants, NamedID{name, id})
	}
	value["landlord"] = landlord
	value["tenants"] = tenants
	value["vendors"] = vendors
	value["managers"] = managers

	return value, nil
}

func SendNotification(text, forUser string) {
	notification := models.Notification{
		Text:       text,
		ReceivedBy: forUser,
	}
	log.Println(
		fmt.Sprintf("Inserting notification %s", notification.Text),
		models.Insert(&notification, models.NotificationCollectionName),
	)
}
