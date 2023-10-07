package middlewares

//func JwtAuth(app *config.Application) fiber.Handler {
//	return func(c *fiber.Ctx) error {
//		user := c.Locals("user").(*jwt.Token)
//		claims := user.Claims.(jwt.MapClaims)
//		username := claims["user"].(string)
//		ctx := context.Background()
//		userObj, err := app.EntClient.User.Query().Where(user2.UsernameEQ(username)).Only(ctx)
//		if err != nil {
//			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
//				"message": "user not found",
//				"error":   err.Error(),
//			})
//		} else {
//			c.Set("userObj", userObj)
//		}
//	}
//}
