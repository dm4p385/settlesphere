package routers

import (
	"github.com/gofiber/fiber/v2"
	"settlesphere/handlers"
)

func SetRoutes(app *fiber.App) {
	api := app.Group("/api/v1/")

	api.Get("/status", handlers.Status)
	//auth := api.Group("/auth")
	//auth.Post("/login")
}
