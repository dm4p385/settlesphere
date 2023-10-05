package routers

import (
	"settlesphere/config"
	"settlesphere/handlers"
)

func SetRoutes(app *config.Application) {
	api := app.FiberApp.Group("/api/v1/")

	api.Get("/status", handlers.Status)
	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login(app))
}
