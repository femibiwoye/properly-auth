package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"properlyauth/database"
	"time"
)

const (
	//UserCollectionName holds the collection on mongodb where user detailas are being stored
	UserCollectionName             = "User"
	phoneNoTempTokenCollectionName = "TempToken"
)

const (
	Manager  = "manager"
	Landlord = "landlord"
	Tenant   = "tenant"
	Vendor   = "vendor"
)

//User decribes user on scoodent
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

//InsertUser insert a user into the database
func InsertUser(user *User) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(UserCollectionName)
	result, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}
	user.ID = result.InsertedID.(primitive.ObjectID).Hex()
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "id", Value: user.ID}}}}
	err = UpdateUser(user, update)
	return err
}

//UpdateUser update a user into the database
func UpdateUser(user *User, update interface{}) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(UserCollectionName)
	s, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": s}

	opts := options.Update().SetUpsert(false)

	_, err = collection.UpdateOne(context.TODO(), filter, update, opts)
	return err
}

//DeleteUser remove a user from the db
func DeleteUser(user *User) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(UserCollectionName)
	s, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": s}

	opts := options.Delete().SetCollation(&options.Collation{
		Locale:    "en_US",
		Strength:  1,
		CaseLevel: false,
	})

	_, err = collection.DeleteOne(context.TODO(), filter, opts)
	return err
}

//SaveToken saves an token  for authentication later on
func SaveToken(key, value, platform string) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	opts := options.Update().SetUpsert(true)
	filter := bson.D{{Key: key}}
	update := bson.D{{"$set", bson.M{"key": key, "value": value, "platform": platform, "time": time.Now().Unix()}}}
	collection := client.Database(database.DbName).Collection(phoneNoTempTokenCollectionName)
	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	return err
}

//FetchToken retrieve the phone and stored token value
func FetchToken(email string) (map[string]interface{}, error) {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(phoneNoTempTokenCollectionName)
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
	collection := client.Database(database.DbName).Collection(phoneNoTempTokenCollectionName)
	filter := bson.M{"key": email}
	opts := options.Delete().SetCollation(&options.Collation{
		Locale:    "en_US",
		Strength:  1,
		CaseLevel: false,
	})
	_, err := collection.DeleteOne(context.TODO(), filter, opts)
	return err
}

//FetchUserByCriterion returns a user struct that tha matches the particular criteria
// i.e FetchUserByCriterion("username","abraham") returns a user struct where username is abraham
func FetchUserByCriterion(criteria, value string) (*User, error) {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(UserCollectionName)
	filter := bson.M{criteria: value}
	user := &User{}

	err := collection.FindOne(context.TODO(), filter).Decode(user)

	if err != nil {
		return nil, err
	}
	return user, nil
}
