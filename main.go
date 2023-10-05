package main

import (
	"github.com/gofiber/fiber/v2"
	"settlesphere/db"
	"settlesphere/routers"
)

func main() {
	db.SetUpEnt()
	fiberApp := fiber.New(fiber.Config{
		ServerHeader: "SettleSphere",
		AppName:      "SettleSphere",
	})
	// setup routes
	routers.SetRoutes(fiberApp)
	fiberApp.Listen(":3000")
}
