package main

import (
	"github.com/gofiber/fiber/v2"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
	"settlesphere/db"
	"settlesphere/routers"
)

func main() {
	db.SetUpEnt()
	fiberApp := fiber.New(fiber.Config{
		ServerHeader: "SettleSphere",
		AppName:      "SettleSphere",
	})
	fiberApp.Use(recover2.New())
	// setup routes
	routers.SetRoutes(fiberApp)
	fiberApp.Listen(":3000")
}
