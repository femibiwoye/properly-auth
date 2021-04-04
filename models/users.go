package models

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"properlyauth/database"
	"time"
)

const (
	//UserCollectionName holds the collection on mongodb where user detailas are being stored
	UserCollectionName      = "User"
	TempTokenCollectionName = "TempToken"
)

const (
	Manager  = "manager"
	Landlord = "landlord"
	Tenant   = "tenant"
	Vendor   = "vendor"
)

//User decribes user on properly
type User struct {
	Email           string `json:"email"`
	FirstName       string `json:"firstname"`
	LastName        string `json:"lastname"`
	ID              string `json:"id"`
	ProfileImageURL string `json:"profile_image_url"`
	Dob             string `json:"dob"`
	CreatedAt       int64  `json:"created_at"`
	PhoneNumber     string `json:"phoneNumber"`
	Password        string `json:"password"`
	Type            string `json:"type"`
	PUMCCode        string `json:"pumccode"`
}

func (u *User) getID() string {
	return u.ID
}

func (u *User) setID(id string) {
	u.ID = id
}

func (u *User) getCreatedAt() int64 {
	return u.CreatedAt
}

func (u *User) setCreatedAt(at int64) {
	u.CreatedAt = at
}

func ToUserFromM(mongoM bson.M) (*User, error) {
	uB, err := bson.Marshal(mongoM)
	if err != nil {
		return nil, err
	}
	u := &User{}
	err = bson.Unmarshal(uB, u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

//SaveToken saves an token  for authentication later on
func SaveToken(key, value, platform string) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	opts := options.Update().SetUpsert(true)
	filter := bson.D{{Key: key}}
	update := bson.D{{"$set", bson.M{"key": key, "value": value, "platform": platform, "time": time.Now().Unix()}}}
	collection := client.Database(database.DbName).Collection(TempTokenCollectionName)
	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	return err
}

//FetchToken retrieve the phone and stored token value
func FetchToken(email string) (map[string]interface{}, error) {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(TempTokenCollectionName)
	filter := bson.M{"key": email}
	res := make(map[string]interface{})
	err := collection.FindOne(context.TODO(), filter).Decode(res)

	if err != nil {
		return nil, err
	}
	return res, nil
}

//TakeOutToken removes the  phone number token out of db
func TakeOutToken(email string) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(TempTokenCollectionName)
	filter := bson.M{"key": email}
	opts := options.Delete().SetCollation(&options.Collation{
		Locale:    "en_US",
		Strength:  1,
		CaseLevel: false,
	})
	_, err := collection.DeleteOne(context.TODO(), filter, opts)
	return err
}

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
