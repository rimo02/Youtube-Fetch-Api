package controllers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rimo02/youtube-api-server/src/api"
	"github.com/rimo02/youtube-api-server/src/config"
	"github.com/rimo02/youtube-api-server/src/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	youtubeOrderBy     = "date"
	youtubeServicetype = "videos"
)

var (
	youtubePublishedafter = time.Date(2024, time.November, 1, 0, 0, 0, 0, time.UTC)
	collection            *mongo.Collection
)

func SetCollection(client *mongo.Client) {
	collection = client.Database("youtube-fetch-api").Collection("youtube-videos")
}

// insert if not exists, update if exist
func bulkInsertToDB(videos []model.Searchapi) error {
	var models []mongo.WriteModel
	for _, item := range videos {
		videoField := bson.M{
			"title":       item.Title,
			"description": item.Description,
			"channelName": item.ChannelName,
			"publishedAt": item.PublishedAt,
		}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(bson.M{"videoID": item.VideoId}).SetUpdate(videoField).SetUpsert(true))
	}
	opt := options.BulkWrite().SetOrdered(false)
	res, err := collection.BulkWrite(context.Background(), models, opt)
	if err != nil {
		log.Printf("Error in bulk write: %v", err)
		return err
	}
	log.Printf("Inserted %v into collection", res.UpsertedIDs)
	return nil
}

func FetchNewVideos(c *fiber.Ctx) error {

	ctx := context.Background()
	key := config.GetApiKey()

	if key == "" {
		var err error
		key, err := api.GetValidApiKey()
		if err != nil {
			log.Fatalf("Error :%v", err)
			return err
		}
		config.SetApiKey(key)
	}

	service, err := youtube.NewService(ctx, option.WithAPIKey(key))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "error in fetching data from the api",
		})
	}

	// Make the API call to YouTube.
	call := service.Search.List([]string{"id,snippet"}).
		Q(config.GetQuery()).
		MaxResults(int64(config.MaxVideosFetched())).
		Order(youtubeOrderBy).
		Type(youtubeServicetype).
		PublishedAfter(youtubePublishedafter.String())

	response, err := call.Do()

	if config.GetEtag() != "" {

	}

	if err != nil {

		if strings.Contains(err.Error(), "quotaExceeded") {
			key, err := api.GetValidApiKey()
			if err != nil {
				log.Printf("Error: %v", err)
				return err
			}
			config.SetApiKey(key)
			return c.Status(http.StatusForbidden).JSON(map[string]string{
				"error": "API quota exceeded",
			})
		}
	}
	videos := make([]model.Searchapi, 0)

	for _, item := range response.Items {
		video := model.Searchapi{
			VideoId:     item.Id.VideoId,
			Title:       item.Snippet.Title,
			Description: item.Snippet.Description,
			ChannelId:   item.Snippet.ChannelId,
			ChannelName: item.Snippet.ChannelTitle,
			PublishedAt: item.Snippet.PublishedAt,
		}
		videos = append(videos, video)
	}
	err = bulkInsertToDB(videos)
	if err != nil {
		log.Printf("FetchNewVideosAndUpdateDb: Error inserting into db: %v", err)
		return err
	}
	config.SetEtag(response.Etag)
	return nil
}
