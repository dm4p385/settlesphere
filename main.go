package main

import (
	"github.com/gofiber/fiber/v2"
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
	app := config.InitializeApp(fiberApp, entClient)
	fiberApp.Use(recover2.New())
	// setup routes
	routers.SetRoutes(app)
	fiberApp.Listen(":3000")
}
