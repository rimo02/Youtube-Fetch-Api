package api

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/rimo02/youtube-api-server/src/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"log"
	"time"
)

var collection *mongo.Collection

func SetCollection(client *mongo.Client) {
	collection = client.Database("youtube-fetch-api").Collection("api-keys")
}
func GetValidApiKey() (string, error) {
	if config.GetApiKey() != "" {
		return config.GetApiKey(), nil
	}
	ctx := context.Background()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error : %v", err)
		return "", err
	}
	for cursor.Next(ctx) {
		var key bson.M
		err := cursor.Decode(&key)
		if err != nil {
			log.Print("Error\n")
			continue
		}
		if key["expired"] == "true" {
			continue
		} else {
			return key["key"].(string), nil
		}
	}
	return "", errors.New("no valid api keys found")
}

func Insert(key string) error {
	_, err := collection.InsertOne(context.TODO(), bson.M{"key": key, "expired": "false", "lasttime": time.Now()})
	if err != nil {
		log.Printf("Error inserting key: %v", err)
		return err
	}
	return nil
}

func IsKeyValid(key string) bool {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(key))
	if err != nil {
		log.Printf("Error creating youtube service : %v", err)
		return false
	}
	call := service.Search.List([]string{"id"}).Q("google")
	_, err1 := call.Do()
	if err1 != nil {
		log.Printf("Invalid API key: %v\n", err1)
		return false
	}
	return true
}

func DeleteExpiredApiKeys(c *fiber.Ctx) {
	ctx := context.Background()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
	}
	for cursor.Next(ctx) {
		var key bson.M
		err := cursor.Decode(&key)
		if err != nil {
			log.Print("Error\n")
			continue
		}
		if time.Since(key["lasttime"].(time.Time)) > 24*time.Hour {
			_, err := collection.DeleteOne(ctx, bson.M{"_id": key["_id"]})
			if err != nil {
				log.Println("Error deleting expired API key:", err)
			} else {
				log.Printf("Deleted expired API key with ID: %v", key["_id"])
			}
		}
	}
}
