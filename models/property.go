package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"properlyauth/database"
)

const (
	//PropertyCollectionName holds the collection for property
	PropertyCollectionName = "Property"
)

//Propertu decribes user property on properly
type Property struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Status    string            `json:"status"`
	Address   string            `json:"address"`
	Images    []string          `json:"images"`
	Documents []string          `json:"documents"`
	Landlords map[string]string `json:"landlord"`
	Tenants   map[string]string `json:"tenants"`
	Vendors   map[string]string `json:"vendors"`
	CreatedAt int64             `json:"created_at"`
	CreatedBy string            `json:"created_by"`
}

//InsertProperty insert a property into the database
func InsertProperty(property *Property) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(PropertyCollectionName)
	result, err := collection.InsertOne(context.TODO(), property)
	if err != nil {
		return err
	}
	property.ID = result.InsertedID.(primitive.ObjectID).Hex()
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "id", Value: property.ID}}}}
	err = UpdateProperty(property, update)
	return err
}

//UpdateProperty update a property into the database
func UpdateProperty(property *Property, update interface{}) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(PropertyCollectionName)
	s, err := primitive.ObjectIDFromHex(property.ID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": s}

	opts := options.Update().SetUpsert(false)

	_, err = collection.UpdateOne(context.TODO(), filter, update, opts)
	return err
}

//DeleteProperty remove a property from the db
func DeleteProperty(user *User) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(PropertyCollectionName)
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

//FetchPropertyByCriterion returns a property struct that matches the particular criteria
// i.e FetchPropertyByCriterion("Name","abraham") returns a user struct where Name is abraham
func FetchPropertyByCriterion(criteria, value string) (*Property, error) {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(PropertyCollectionName)
	filter := bson.M{criteria: value}
	property := &Property{}

	err := collection.FindOne(context.TODO(), filter).Decode(property)

	if err != nil {
		return nil, err
	}
	return property, nil
}
