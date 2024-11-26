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
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-ticker.C:
				ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
				err := controllers.FetchNewVideos(ctx)
				if err != nil {
					log.Errorf("Error fetching videos and updating database %v", err)
					return
				}
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
			}
		}
	}()
	routes.SetSearchRoutes(app)
	app.Listen(":3000")
}
