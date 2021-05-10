package database

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	//DbName stands for the entire DB name
	DbName = "properly"
)

//DB returns mongodb connecter struct
type DB struct {
	client     *mongo.Client
	ctx        *context.Context
	cancelFunc context.CancelFunc
}

// Done cancels these DB connection
func (db *DB) Done() {
	db.cancelFunc()
}

//GetClient returns the mongo db connected client
func (db *DB) GetClient() *mongo.Client {
	return db.client
}
func newDB() *DB {
	mongourl := os.Getenv("MONGO_URL")
	if len(mongourl) <= 0 {
		mongourl = "mongodb://localhost:27017/"
	}
	dbName := os.Getenv("DBNAME")
	if len(dbName) > 0 {
		DbName = dbName
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(mongourl))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}
	return &DB{client: client, ctx: &ctx, cancelFunc: cancelFunc}
}

//MongoDBPool objects
var mongoDBPool = sync.Pool{
	New: func() interface{} {
		return newDB()
	},
}

//GetMongoDB returns  a DB instnace from a pool of them
func GetMongoDB() *DB {
	db := mongoDBPool.Get().(*DB)
	return db
}

//PutDBBack returns a used object back to the pool
func PutDBBack(db *DB) {
	mongoDBPool.Put(db)
}
