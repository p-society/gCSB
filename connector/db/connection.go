package generaldb

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Collection *mongo.Collection
var Database *mongo.Database
var Client *mongo.Client

func Init() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	fmt.Println("Env Loaded Successfully.")

	// Access environment variables
	dbLink := os.Getenv("MONGO_URI")
	const databaseName = "generaldb"
	const collection1Name = "all-players"

	clientOptions := options.Client().ApplyURI(dbLink)
	Client, err = mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal("Error Occurred while connecting to the database:", err)
	}

	fmt.Println("Connection to Database Successful")

	Database = Client.Database(databaseName)
	Collection = Database.Collection(collection1Name)
	fmt.Printf("Collection Instance %s is Ready.", collection1Name)
}

func ProvideAptCollection(dbName, colName string) *mongo.Collection {
	return Client.Database(dbName).Collection(colName)
}
