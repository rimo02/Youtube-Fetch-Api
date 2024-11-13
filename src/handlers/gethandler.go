package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rimo02/youtube-api-server/src/api"
	"github.com/rimo02/youtube-api-server/src/controllers"
)

func DoGetVid(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "the page you are asking for does not exist",
		})
	}
	videos := controllers.GetVideos(int64(page))
	return c.JSON(
		fiber.Map{
			"Videos": videos,
		},
	)
}

func DoSetApi(c *fiber.Ctx) error {
	apiKey := c.Query("key", "")
	if apiKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "please provide an api key",
		})
	}
	err := api.Insert(apiKey)
	if !api.IsKeyValid(apiKey) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid api key",
		})
	}
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "error inserting key to the database",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "api key successfully added",
	})
}
