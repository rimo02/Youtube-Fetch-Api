package db

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rimo02/youtube-api-server/src/api"
	"github.com/rimo02/youtube-api-server/src/controllers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

var (
	Client *mongo.Client
)

func connectDB() *mongo.Client {
	godotenv.Load()
	var uri = os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(uri)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB: ", err)
	}

	fmt.Printf("Connected to MongoDB")
	return client
}
func InitMongoDB() {
	Client := connectDB()
	controllers.SetCollection(Client)
	api.SetCollection(Client)
}
