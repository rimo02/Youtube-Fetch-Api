package controllers

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rimo02/youtube-api-server/src/api"
	"github.com/rimo02/youtube-api-server/src/config"
	"github.com/rimo02/youtube-api-server/src/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"log"
	"net/http"
	"strings"
	"time"
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

func getVideoFromDB(videoId string) (model.Searchapi, error) {
	var video model.Searchapi
	err := collection.FindOne(context.Background(), bson.M{"videoId": videoId}).Decode(&video)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Searchapi{}, nil // Video does not exist
		}
		return model.Searchapi{}, err // Some other error
	}
	return video, nil
}

// insert if not exists, update if exist
func bulkInsertToDB(videos []model.Searchapi) error {
	var models []mongo.WriteModel
	for _, item := range videos {

		storedVideo, err := getVideoFromDB(item.VideoId)
		if err != nil {
			log.Printf("Error fetching video from DB: %v", err)
			continue
		}

		// if video does not exist in the database or the etag is different, prepare an update
		if storedVideo.Etag != item.Etag || storedVideo.Etag == "" {

			videoField := bson.M{
				"$set": bson.M{
					"videoId":     item.VideoId,
					"title":       item.Title,
					"description": item.Description,
					"channelId":   item.ChannelId,
					"channelName": item.ChannelName,
					"publishedAt": item.PublishedAt,
					"etag":        item.Etag,
				},
			}
			models = append(models,
				mongo.NewUpdateOneModel().
					SetFilter(bson.M{"videoId": item.VideoId}).
					SetUpdate(videoField).
					SetUpsert(true),
			)
		}
	}

	if len(models) > 0 {
		opt := options.BulkWrite().SetOrdered(false)
		res, err := collection.BulkWrite(context.Background(), models, opt)
		if err != nil {
			log.Printf("Error in bulk write: %v", err)
			return err
		}
		log.Printf("Bulk write result - Inserted: %d, Modified: %d",
			len(res.UpsertedIDs),
			res.ModifiedCount,
		)
	}
	return nil
}

func FetchNewVideos(c *fiber.Ctx) error {

	ctx := context.Background()
	key := config.GetApiKey()

	if key == "" {
		var err error
		key, err := api.GetValidApiKey()
		if err != nil {
			log.Printf("Error getting API key: %v", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve API key",
			})
		}
		config.SetApiKey(key)
	}

	service, err := youtube.NewService(ctx, option.WithAPIKey(config.GetApiKey()))
	if err != nil {
		log.Printf("Error creating YouTube service: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Error creating YouTube service",
		})
	}

	// Make the API call to YouTube.
	call := service.Search.List([]string{"id,snippet"}).
		Q(config.GetQuery()).
		MaxResults(int64(config.MaxVideosFetched())).
		Order(youtubeOrderBy).
		Type(youtubeServicetype).
		PublishedAfter(youtubePublishedafter.Format(time.RFC3339))

	if config.GetEtag() != "" { // etag has already been added
		call = call.IfNoneMatch(config.GetEtag())
	}

	response, err := call.Do()
	if err != nil {

		if strings.Contains(err.Error(), "quotaExceeded") {
			key, err := api.GetValidApiKey()
			if err != nil {
				log.Printf("Error getting new API key: %v", err)
				return c.Status(http.StatusForbidden).JSON(fiber.Map{
					"error": "API quota exceeded and unable to get new key",
				})
			}
			config.SetApiKey(key)
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": "API quota exceeded. New key retrieved.",
			})
		}
		log.Printf("Error fetching YouTube videos: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch videos",
		})
	}

	if response.Etag == config.GetEtag() {
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"message": "no data has been changed. Skipping insertion",
		})
	}
	videos := make([]model.Searchapi, 0, len(response.Items))

	for _, item := range response.Items {
		video := model.Searchapi{
			VideoId:     item.Id.VideoId,
			Title:       item.Snippet.Title,
			Description: item.Snippet.Description,
			ChannelId:   item.Snippet.ChannelId,
			ChannelName: item.Snippet.ChannelTitle,
			PublishedAt: item.Snippet.PublishedAt,
			Etag:        item.Etag,
		}
		videos = append(videos, video)
	}
	if err := bulkInsertToDB(videos); err != nil {
		log.Printf("Error inserting videos into DB: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store videos in database",
		})
	}
	config.SetEtag(response.Etag)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("Successfully fetched and stored %d videos", len(videos)),
		"count":   len(videos),
	})
}
