package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/rimo02/youtube-api-server/db"
	"github.com/rimo02/youtube-api-server/src/api"
	"github.com/rimo02/youtube-api-server/src/config"
	"github.com/rimo02/youtube-api-server/src/controllers"
	"github.com/rimo02/youtube-api-server/src/routes"
	"github.com/valyala/fasthttp"
	"time"
)

func init() {
	db.InitMongoDB()
	config.InitConfig()
}

func main() {
	// start a goroutine to fetch videos from youtube periodically
	app := fiber.New()

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		for {
			select {
			case <-ticker.C:
				ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
				err := controllers.FetchNewVideos(ctx)
				if err != nil {
					log.Errorf("Error fetching videos and updating database %v", err)
					return
				}
			case <-done:
				log.Info("Stopping video fetcher goroutine...")
				return
			}
		}
	}()

	// goroutine to delete expired api keys
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for {
			select {
			case <-ticker.C:
				ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
				api.DeleteExpiredApiKeys(ctx)
			case <-done:
				log.Info("Stopping expired API key deleter goroutine...")
				return
			}
		}
	}()
	routes.SetSearchRoutes(app)
	app.Listen(":3000")
}
