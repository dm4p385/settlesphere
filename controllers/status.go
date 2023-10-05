package controllers

import "github.com/gofiber/fiber/v2"

func Status(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"Status": "Server is healthy!",
	})
}
