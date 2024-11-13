package controllers

import (
	"context"
	"log"
	"time"
	"github.com/rimo02/youtube-api-server/src/config"
	"github.com/rimo02/youtube-api-server/src/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// get videos from the database in a paginated format
func GetVideos(page int64) []model.Searchapi {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// searching videos in a paginated formar using skip and limit
	opt := options.Find()
	opt.SetSkip((page - 1) * int64(config.LimitPerPage())) // skips (n-1)* limits collections
	opt.SetLimit(int64(config.LimitPerPage()))

	cursor, err := collection.Find(ctx, bson.M{}, opt)
	if err != nil {
		log.Printf("GetVideos: Error fetching videos: %v\n", err)
		return nil
	}
	videos := make([]model.Searchapi, 0)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var video model.Searchapi
		err := cursor.Decode(&video)
		if err != nil {
			log.Printf("Error in decoding video: %s", err)
			continue
		}
		videos = append(videos, video)
	}
	return videos
}
