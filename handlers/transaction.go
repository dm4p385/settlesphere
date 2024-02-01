package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"settlesphere/config"
	"settlesphere/ent"
	"settlesphere/ent/group"
	"settlesphere/ent/user"
	"settlesphere/services"
	"strconv"
	"time"
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
		txn, err := userOps.GetUserTxns(userObj, groupObj)
		if err != nil {
			log.Error(err)
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
		// Lender is the one who owes money
		// Receiver lends the money
		req := struct {
			Receiver int                `json:"receiver"`
			Lender   map[string]float64 `json:"lender"`
			Amount   float64            `json:"amount"`
			Note     string             `json:"note"`
		}{}
		if err := c.BodyParser(&req); err != nil {
			log.Error(err)
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
		receiver, err := app.EntClient.User.Query().Where(user.IDEQ(req.Receiver)).Only(ctx)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "receiver does not exist",
				"error":   err.Error(),
			})
		}

		// i don't like having two loops for this but I don't think I have a lot of choice here
		//for receiverId, receiverAmount := range req.Receiver {
		//	_, err := app.EntClient.User.Query().Where(user.IDEQ(receiverId)).Only(ctx)
		//	if err != nil {
		//		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		//			"message": "receiver does not exist",
		//			"error": err.Error(),
		//		})
		//	}
		//	if receiverAmount <= 0 {
		//		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		//			"message": "amount cannot be negative or zero",
		//		})
		//	}
		//}

		txnOps := services.NewTxnOps(ctx, app)
		var txnArray []*ent.Transaction
		// this method is bad, the transaction gets termination in between instead of being all or nothing
		for lenderIdString, lenderAmount := range req.Lender {
			lenderId, err := strconv.Atoi(lenderIdString)
			lender, err := app.EntClient.User.Query().Where(user.IDEQ(lenderId)).Only(ctx)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"message": "lender does not exist",
					"error":   err.Error(),
				})
			}
			if lenderAmount <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "amount cannot be negative or zero",
				})
			}
			if receiver.ID != lenderId {
				txn, err := txnOps.GenerateTransaction(groupObj, lender, receiver, lenderAmount, req.Note, req.Amount)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"message": "something went wrong",
						"error":   err.Error(),
					})
				}
				txnArray = append(txnArray, txn)
			}

			err = userOps.UpdateUserShareStat(req.Lender[strconv.Itoa(lenderId)], lender, groupObj)
			if err != nil {
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"message": "something went wrong while updating user stats",
						"error":   err.Error(),
					})
				}
			}

		}

		err = userOps.UpdateUserPaidByStat(req.Amount, receiver, groupObj)
		if err != nil {
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "something went wrong while updating user stats",
					"error":   err.Error(),
				})
			}
		}

		//receiver, err := app.EntClient.User.Query().Where(user.IDEQ(req.Receiver)).Only(ctx)
		//if err != nil {
		//	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		//		"message": "receiver does not exist",
		//		"error": err.Error(),
		//	})
		//}
		//if req.Amount <= 0 {
		//	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		//		"message": "amount cannot be negative or zero",
		//	})
		//}
		//txnOps := services.NewTxnOps(ctx, app)
		//txn, err := txnOps.GenerateTransaction(groupObj, lender, receiver, req.Amount, req.Note)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "something went wrong",
				"error":   err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "transaction created successfully",
			"txn":     txnArray,
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
		groupOps := services.NewGroupOps(ctx, app)
		txnHistoryObjArr, err := groupOps.GetTxnHistoryOfGroup(groupObj)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "something went wrong",
				"error":   err.Error(),
			})
		}
		type txnHistoryRes struct {
			TxnId       int        `json:"id"`
			Note        string     `json:"note"`
			ReceiverId  int        `json:"receiverId"`
			PayerId     int        `json:"payerId"`
			Amount      float64    `json:"amount"`
			TotalAmount float64    `json:"total_amount"`
			Settled     bool       `json:"settled"`
			CreatedAt   time.Time  `json:"created_at"`
			SettledAt   *time.Time `json:"settled_at,omitempty"`
		}
		var txnHistoryArr []txnHistoryRes
		for _, txnHistory := range txnHistoryObjArr {
			//var settleTime time.Time
			//if txnHistory.SettledAt != nil {
			//	settleTime = *txnHistory.SettledAt
			//}
			temp := txnHistoryRes{
				TxnId:       txnHistory.ID,
				Note:        txnHistory.Note,
				ReceiverId:  txnHistory.QueryDestination().OnlyIDX(ctx),
				PayerId:     txnHistory.QuerySource().OnlyIDX(ctx),
				Amount:      txnHistory.Amount,
				TotalAmount: txnHistory.TotalAmount,
				Settled:     txnHistory.Settled,
				CreatedAt:   txnHistory.CreatedAt,
				SettledAt:   txnHistory.SettledAt,
			}
			txnHistoryArr = append(txnHistoryArr, temp)
		}
		netAmount, err := groupOps.GetUserNetAmountOfGroup(userObj, groupObj)
		if err != nil {
			netAmount = 0
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":     "transaction history",
			"txn_history": txnHistoryArr,
			"netAmount":   netAmount,
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
		targetUserId, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "failed to parse user ID",
				"error":   err.Error(),
			})
		}
		targetUserObj, err := app.EntClient.User.Query().Where(user.IDEQ(targetUserId)).Only(ctx)
		if err != nil {
			log.Error(err)
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
			log.Error(err)
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
