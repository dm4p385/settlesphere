package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"settlesphere/ent/group"
	"settlesphere/services"

	"settlesphere/config"
)

func ListGroups(app *config.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		userOps := services.NewUserOps(ctx, app)
		token := c.Locals("user").(*jwt.Token)
		userObj, err := userOps.GetUserByJwt(token)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "user not found",
				"error":   err.Error(),
			})
		}
		type groupsRes struct {
			Name string `json:"name"`
			Code string `json:"code"`
		}

		// TODO: fix this response
		groups, err := userObj.QueryMemberOf().Select(group.FieldName).Select(group.FieldCode).Select(group.FieldCreatedBy).All(ctx)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "something went wrong",
				"error":   err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "fetched groups",
			"groups":  groups,
		})
	}
}

func JoinGroup(app *config.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := struct {
			GroupCodeString string `json:"group_code"`
		}{}
		if err := c.BodyParser(&req); err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "the request is not in the correct format",
				"error":   err,
			})
		}
		groupCode, err := uuid.Parse(req.GroupCodeString)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid group code",
				"error":   err.Error(),
			})
		}
		ctx := context.Background()
		group, err := app.EntClient.Group.Query().Where(group.CodeEQ(groupCode)).Only(ctx)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "group not found",
				"error":   err.Error(),
			})
		}
		userOps := services.NewUserOps(ctx, app)
		token := c.Locals("user").(*jwt.Token)
		userObj, err := userOps.GetUserByJwt(token)
		groupOps := services.NewGroupOps(ctx, app)
		groupOps.AddUserToGroup(group, userObj)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "joined group",
		})
	}
}

func CreateGroup(app *config.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := struct {
			Name string `json:"name"`
		}{}
		if err := c.BodyParser(&req); err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "the request is not in the correct format",
				"error":   err,
			})
		}
		ctx := context.Background()
		userOps := services.NewUserOps(ctx, app)
		token := c.Locals("user").(*jwt.Token)
		userObj, err := userOps.GetUserByJwt(token)
		group, err := app.EntClient.Group.Create().
			SetName(req.Name).
			SetCreatedBy(userObj.Username).
			Save(ctx)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "something went wrong",
				"error":   err.Error(),
			})
		}
		groupOps := services.NewGroupOps(ctx, app)
		groupOps.AddUserToGroup(group, userObj)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "group created",
			"group": fiber.Map{
				"name": group.Name,
				"code": group.Code,
			},
		})
	}
}

func GetUsers(app *config.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		userOps := services.NewUserOps(ctx, app)
		token := c.Locals("user").(*jwt.Token)
		userObj, err := userOps.GetUserByJwt(token)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "user not found",
				"error":   err.Error(),
			})
		}
		groupCodeString := c.Params("code")
		groupCode, err := uuid.Parse(groupCodeString)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid group code",
				"error":   err.Error(),
			})
		}
		groupObj, err := app.EntClient.Group.Query().Where(group.CodeEQ(groupCode)).Only(ctx)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "group not found",
				"error":   err.Error(),
			})
		}
		if isMember, _ := userObj.QueryMemberOf().Where(group.IDEQ(groupObj.ID)).Exist(ctx); !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "user does not belong to this group",
			})
		}
		users, err := groupObj.QueryUsers().All(ctx)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "something went wrong",
				"error":   err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "fetched users",
			"users":   users,
		})
	}
}

func GetSettledTxns(app *config.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		userOps := services.NewUserOps(ctx, app)
		token := c.Locals("user").(*jwt.Token)
		userObj, err := userOps.GetUserByJwt(token)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "user not found",
				"error":   err.Error(),
			})
		}
		groupOps := services.NewGroupOps(ctx, app)
		owedTxnHistory, lentTxnHistory, err := groupOps.GetSettledTxnsOfAllGroups(userObj)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "something went wrong",
				"error":   err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"owed": owedTxnHistory,
			"lent": lentTxnHistory,
		})
	}
}
