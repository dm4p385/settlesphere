package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
	"settlesphere/config"
	"settlesphere/db"
	"settlesphere/routers"
)

func main() {
	entClient := db.SetUpEnt()
	defer entClient.Close()
	fiberApp := fiber.New(fiber.Config{
		ServerHeader: "SettleSphere",
		AppName:      "SettleSphere",
	})
	fiberApp.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin",
		AllowOrigins:     "*",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))
	app := config.InitializeApp(fiberApp, entClient)
	fiberApp.Use(recover2.New())
	// setup routes
	routers.SetRoutes(app)
	fiberApp.Listen(":3000")
}
