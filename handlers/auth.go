package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"settlesphere/config"
	user2 "settlesphere/ent/user"
	"strings"
	"time"
)

func Login(app *config.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := struct {
			Email  string `json:"email"`
			PubKey string `json:"pubKey"`
		}{}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "the request is not in the correct format",
				"error":   err,
			})
		}
		ctx := context.Background()
		user, err := app.EntClient.User.Query().Where(user2.EmailEQ(req.Email)).Only(ctx)
		if err != nil {
			username := strings.Split(req.Email, "@")[0]
			user, err = app.EntClient.User.Create().
				SetUsername(username).
				SetEmail(req.Email).
				SetPubKey(req.PubKey).
				Save(ctx)
			if err != nil {
				log.Errorf("something went wrong while creating a user: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "something went wrong while creating a user",
					"error":   err,
				})
			}
		}
		if user.PubKey != req.PubKey {
			return c.Status(401).JSON(fiber.Map{
				"message": "wrong pubKey for this user",
			})
		}
		// claims
		claims := jwt.MapClaims{
			"user":  user.Username,
			"email": user.Email,
			"exp":   time.Now().Add(time.Hour * 72).Unix(),
		}
		// Create token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte(app.Secret))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "something went wrong while generating JWT token",
				"error":   err,
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "logged in successfully",
			"token":   t,
		})
	}
}
