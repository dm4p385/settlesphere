package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"settlesphere/config"
	"settlesphere/ent/group"
	"settlesphere/ent/user"
	"settlesphere/services"
	"strconv"
)

//func ListTxns(app *config.Application) fiber.Handler {
//	return func(c *fiber.Ctx) error {
//
//	}
//}

func GroupUserTxns(app *config.Application) fiber.Handler {
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
		txn, err := userOps.GetUserTxns(userObj, groupObj)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "something went wrong",
				"error":   err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "user transactions",
			"txns":    txn,
		})
	}
}

func AddTransaction(app *config.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := struct {
			Lender   int    `json:"lender"`
			Receiver int    `json:"receiver"`
			Amount   int    `json:"amount"`
			Note     string `json:"note"`
		}{}
		if err := c.BodyParser(&req); err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "the request is not in the correct format",
				"error":   err.Error(),
			})
		}
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
		lender, err := app.EntClient.User.Query().Where(user.IDEQ(req.Lender)).Only(ctx)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "lender does not exist",
				"error":   err.Error(),
			})
		}
		receiver, err := app.EntClient.User.Query().Where(user.IDEQ(req.Receiver)).Only(ctx)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "receiver does not exist",
				"error":   err.Error(),
			})
		}
		if req.Amount <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "amount cannot be negative or zero",
			})
		}
		txnOps := services.NewTxnOps(ctx, app)
		txn, err := txnOps.GenerateTransaction(groupObj, lender, receiver, req.Amount, req.Note)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "something went wrong",
				"error":   err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "transaction created successfully",
			"txn":     txn,
		})
	}
}

func TxnHistory(app *config.Application) fiber.Handler {
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
		groupOps := services.NewGroupOps(ctx, app)
		txnHistoryObjArr, err := groupOps.GetTxnHistoryOfGroup(groupObj)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "something went wrong",
				"error":   err.Error(),
			})
		}
		type txnHistoryRes struct {
			TxnId      int    `json:id`
			Note       string `json:"note"`
			ReceiverId int    `json:"receiverId"`
			PayerId    int    `json:"payerId"`
			Amount     int    `json:"amount"`
			Settled    bool   `json:"settled"`
		}
		var txnHistoryArr []txnHistoryRes
		for _, txnHistory := range txnHistoryObjArr {
			temp := txnHistoryRes{
				TxnId:      txnHistory.ID,
				Note:       txnHistory.Note,
				ReceiverId: txnHistory.QueryDestination().OnlyIDX(ctx),
				PayerId:    txnHistory.QuerySource().OnlyIDX(ctx),
				Amount:     txnHistory.Amount,
				Settled:    txnHistory.Settled,
			}
			txnHistoryArr = append(txnHistoryArr, temp)
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":     "transaction history",
			"txn_history": txnHistoryArr,
		})
	}
}

func SettleTxn(app *config.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		userOps := services.NewUserOps(ctx, app)
		token := c.Locals("user").(*jwt.Token)
		userObj, err := userOps.GetUserByJwt(token)
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
		targetUserId, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "failed to parse user ID",
				"error":   err.Error(),
			})
		}
		targetUserObj, err := app.EntClient.User.Query().Where(user.IDEQ(targetUserId)).Only(ctx)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "failed to find the user",
				"error":   err.Error(),
			})
		}
		if (!userObj.QueryMemberOf().Where(group.IDEQ(groupObj.ID)).ExistX(ctx)) && (!targetUserObj.QueryMemberOf().Where(group.ID(groupObj.ID)).ExistX(ctx)) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "user does not belong to this group",
			})
		}
		txn, userInfoTxn, err := userOps.SettleTxn(userObj, targetUserObj, groupObj)
		if err != nil {
			log.Errorf(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "failed to settle transaction",
				"error":   err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":     "outstanding transaction successfully settled",
			"txn_history": txn,
			"user_info":   userInfoTxn,
		})
	}
}
