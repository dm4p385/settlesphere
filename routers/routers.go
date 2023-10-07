package routers

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"settlesphere/config"
	"settlesphere/handlers"
)

func SetRoutes(app *config.Application) {
	api := app.FiberApp.Group("/api/v1/")

	api.Get("/status", handlers.Status)

	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login(app))

	group := api.Group("/groups")
	group.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(app.Secret)},
	}))
	group.Get("/", handlers.ListGroups(app))
	group.Get("/join/:code", handlers.JoinGroup(app))
	group.Post("/", handlers.CreateGroup(app))
	//group.Post("/groups")
}
