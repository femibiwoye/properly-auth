package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"properlyauth/database"
)

type ProperlyDocModel interface {
	getID() string
	setID(id string)
	setCreatedAt(at int64)
	getCreatedAt() int64
}

//InsertUser insert a user into the database
func Insert(doc ProperlyDocModel, collectionName string) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	result, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		return err
	}
	doc.setID(result.InsertedID.(primitive.ObjectID).Hex())
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "id", Value: doc.getID()}}}}
	err = Update(doc, update, collectionName)
	return err
}

//Update update a doc in the database
func Update(doc ProperlyDocModel, update interface{}, collectionName string) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	s, err := primitive.ObjectIDFromHex(doc.getID())
	if err != nil {
		return err
	}
	filter := bson.M{"_id": s}

	opts := options.Update().SetUpsert(false)

	_, err = collection.UpdateOne(context.TODO(), filter, update, opts)
	return err
}

//Deleter remove a doc from the db
func Delete(doc ProperlyDocModel, collectionName string) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	s, err := primitive.ObjectIDFromHex(doc.getID())
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

//FetchDocByCriterion returns a  struct that tha matches the particular criteria
// i.e FetchDocByCriterion("username","abraham","user") returns a user struct where username is abraham
func FetchDocByCriterion(criteria, value, collectionName string) (bson.M, error) {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	filter := bson.M{criteria: value}
	doc := bson.M{}

	err := collection.FindOne(context.TODO(), filter).Decode(doc)

	if err != nil {
		return nil, err
	}
	return doc, nil
}

//FetchDocByCriterionMultiple returns a doc struct that tha matches the particular criteria
// i.e FetchDocByCriterionMultiple("username","abraham") returns a user struct where username is abraham and more
func FetchDocByCriterionMultiple(criteria, collectionName string, values []string) ([]bson.M, error) {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	filter := bson.M{criteria: bson.M{"$in": values}}
	users := []bson.M{}

	opts := options.Find().SetSort(bson.D{{"CreatedAt", 1}}).SetProjection(bson.M{"password": 0})
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &users); err != nil {
		return nil, err
	}
	return users, nil
}
