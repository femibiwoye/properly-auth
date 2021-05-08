package models

import (
	"context"
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
	Email           string                 `json:"email"`
	FirstName       string                 `json:"firstname"`
	LastName        string                 `json:"lastname"`
	ID              string                 `json:"id"`
	ProfileImageURL string                 `json:"profile_image_url"`
	Dob             string                 `json:"dob"`
	CreatedAt       int64                  `json:"created_at"`
	PhoneNumber     string                 `json:"phoneNumber"`
	Password        string                 `json:"password"`
	Type            string                 `json:"type"`
	PUMCCode        string                 `json:"pumccode"`
	AdditionalData  map[string]interface{} `json:"additionaldata"`
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

func GetUser(criteria, value string) (*User, error) {
	m, err := FetchDocByCriterion(criteria, value, UserCollectionName)
	if err != nil {
		return nil, err
	}
	return ToUserFromM(m)
}

//SaveToken saves an token  for authentication later on
func SaveToken(key, value, platform string) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	opts := options.Update().SetUpsert(true)
	filter := bson.D{{Key: key}}
	update := bson.M{"$set": bson.M{
		"key":      key,
		"value":    value,
		"platform": platform,
		"time":     time.Now().Unix(),
	},
	}
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
