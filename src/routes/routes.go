package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rimo02/youtube-api-server/src/handlers"
)

var SetSearchRoutes = func(app *fiber.App) {
	app.Get("/api/get_videos", handlers.DoGetVid) //getting videos in paginated way from the database
	app.Get("/api/set_api", handlers.DoSetApi)    //add new api keys to the database
}
