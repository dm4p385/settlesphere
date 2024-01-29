package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"settlesphere/ent/group"
	"settlesphere/services"
	"time"

	"settlesphere/config"
)

func ListGroups(app *config.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		userOps := services.NewUserOps(ctx, app)
		groupOps := services.NewGroupOps(ctx, app)
		token := c.Locals("user").(*jwt.Token)
		userObj, err := userOps.GetUserByJwt(token)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "user not found",
				"error":   err.Error(),
			})
		}
		type groupsRes struct {
			Name      string    `json:"name"`
			Code      uuid.UUID `json:"code"`
			CreatedBy string    `json:"created_by"`
			CreatedAt time.Time `json:"created_at"`
			Image     string    `json:"image"`
			NetAmount float64   `json:"net_amount"`
		}

		// TODO: fix this response
		groups, err := userObj.QueryMemberOf().Select(group.FieldName).Select(group.FieldCode).Select(group.FieldCreatedBy).Select(group.FieldCreatedAt).Select(group.FieldImage).All(ctx)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "something went wrong",
				"error":   err.Error(),
			})
		}
		var resGroups []groupsRes
		for _, groupObj := range groups {
			// netAmount = lent - owed
			// negative implies you owe money
			// positive implies you are owed money
			netAmount, err := groupOps.GetUserNetAmountOfGroup(userObj, groupObj)
			if err != nil {
				netAmount = 0
			}
			resGroups = append(resGroups, groupsRes{
				Name:      groupObj.Name,
				Code:      groupObj.Code,
				CreatedBy: groupObj.CreatedBy,
				CreatedAt: groupObj.CreatedAt,
				Image:     groupObj.Image,
				NetAmount: netAmount,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "fetched groups",
			"groups":  resGroups,
		})
	}
}

func JoinGroup(app *config.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := struct {
			GroupCodeString string `json:"group_code"`
		}{}
		if err := c.BodyParser(&req); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "the request is not in the correct format",
				"error":   err,
			})
		}
		groupCode, err := uuid.Parse(req.GroupCodeString)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid group code",
				"error":   err.Error(),
			})
		}
		ctx := context.Background()
		group, err := app.EntClient.Group.Query().Where(group.CodeEQ(groupCode)).Only(ctx)
		if err != nil {
			log.Error(err)
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
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "the request is not in the correct format",
				"error":   err,
			})
		}
		form, err := c.MultipartForm()
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "could not parse image",
				"error":   err,
			})
		}
		log.Debug(form.File["image"])
		file := form.File["image"][0]
		log.Debugf(file.Filename, file.Size, file.Header["Content-Type"][0])
		url, err := services.UploadToFirebase(app.FirebaseStorageClient, file)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "failed to upload image to firebase",
				"error":   err,
			})
		}
		//for _, file := range form.File["image"] {
		//	log.Debugf(file.Filename, file.Size, file.Header["Content-Type"][0])
		//	url, err := services.UploadToFirebase(app.FirebaseStorageClient, file)
		//	if err != nil {
		//		log.Error(err)
		//		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		//			"message": "failed to upload image to firebase",
		//			"error":   err,
		//		})
		//	}
		//}
		//imageFile := form.File["image"]
		//log.Debug(imageFile.Filename, imageFile.Size, file.Header["Content-Type"][0])
		ctx := context.Background()
		userOps := services.NewUserOps(ctx, app)
		token := c.Locals("user").(*jwt.Token)
		userObj, err := userOps.GetUserByJwt(token)
		group, err := app.EntClient.Group.Create().
			SetName(req.Name).
			SetCreatedBy(userObj.Username).
			SetImage(url).
			Save(ctx)
		if err != nil {
			log.Error(err)
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
				"name":  group.Name,
				"code":  group.Code,
				"image": group.Image,
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
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "user not found",
				"error":   err.Error(),
			})
		}
		groupCodeString := c.Params("code")
		groupCode, err := uuid.Parse(groupCodeString)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid group code",
				"error":   err.Error(),
			})
		}
		groupObj, err := app.EntClient.Group.Query().Where(group.CodeEQ(groupCode)).Only(ctx)
		if err != nil {
			log.Error(err)
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
			log.Error(err)
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
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "user not found",
				"error":   err.Error(),
			})
		}
		groupOps := services.NewGroupOps(ctx, app)
		owedTxnHistory, lentTxnHistory, err := groupOps.GetSettledTxnsOfAllGroups(userObj)
		if err != nil {
			log.Error(err)
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

func GetGroupStats(app *config.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"total_group_spending": 36782.23,
			"total_you_paid_for":   36782.23,
			"your_total_share":     36782.23,
		})
	}
}
