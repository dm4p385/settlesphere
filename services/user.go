package services

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"settlesphere/config"
	"settlesphere/ent"
	"settlesphere/ent/group"
	"settlesphere/ent/transaction"
	user2 "settlesphere/ent/user"
)

type UserOps struct {
	ctx context.Context
	app *config.Application
}

func NewUserOps(ctx context.Context, app *config.Application) *UserOps {
	return &UserOps{
		ctx: ctx,
		app: app,
	}
}

func (r *UserOps) GetUserByJwt(token *jwt.Token) (*ent.User, error) {
	claims := token.Claims.(jwt.MapClaims)
	ctx := context.Background()
	userObj, err := r.app.EntClient.User.Query().Where(user2.UsernameEQ(claims["user"].(string))).Only(ctx)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return userObj, nil
}

type txn struct {
	Owes     []*ent.Transaction `json:"owes"`
	Receives []*ent.Transaction `json:"receives"`
}

func (r *UserOps) GetUserTxns(user *ent.User, groupObj *ent.Group) (txn, error) {
	lentTxns, err := r.app.EntClient.Transaction.Query().
		Where(
			transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
			transaction.HasDestinationWith(user2.IDEQ(user.ID)),
		).All(r.ctx)
	if err != nil {
		return txn{}, err
	}
	owedTxns, err := r.app.EntClient.Transaction.Query().
		Where(
			transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
			transaction.HasSourceWith(user2.IDEQ(user.ID)),
		).All(r.ctx)
	if err != nil {
		return txn{}, err
	}
	txn := txn{
		Owes:     owedTxns,
		Receives: lentTxns,
	}
	return txn, nil
}

// SettleTxn : Its redundant, this functionality should be incorporated in Generate Transaction only
func (r *UserOps) SettleTxn(settler *ent.User, targetUser *ent.User, groupObj *ent.Group) (*ent.TxnHistory, error) {
	existingLentTxn := 0
	existingOwedTxn := 0
	var err error
	if temp := r.app.EntClient.Transaction.Query().
		Where(
			transaction.And(
				transaction.HasSourceWith(user2.IDEQ(settler.ID)),
				transaction.HasDestinationWith(user2.IDEQ(targetUser.ID)),
			),
			transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
		).ExistX(r.ctx); temp {
		existingLentTxn, err = r.app.EntClient.Transaction.Query().
			Where(
				transaction.And(
					transaction.HasSourceWith(user2.IDEQ(settler.ID)),
					transaction.HasDestinationWith(user2.IDEQ(targetUser.ID)),
				),
				transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
			).Aggregate(ent.Sum(transaction.FieldAmount)).Int(r.ctx)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	if temp := r.app.EntClient.Transaction.Query().
		Where(
			transaction.And(
				transaction.HasDestinationWith(user2.IDEQ(settler.ID)),
				transaction.HasSourceWith(user2.IDEQ(targetUser.ID)),
			),
			transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
		).ExistX(r.ctx); temp {
		existingOwedTxn, err = r.app.EntClient.Transaction.Query().
			Where(
				transaction.And(
					transaction.HasDestinationWith(user2.IDEQ(settler.ID)),
					transaction.HasSourceWith(user2.IDEQ(targetUser.ID)),
				),
				transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
			).Aggregate(ent.Sum(transaction.FieldAmount)).Int(r.ctx)
		if err != nil {
			return nil, err
		}
	}
	_, err =
		r.app.EntClient.Transaction.Delete().
			Where(
				transaction.Or(
					transaction.And(
						transaction.HasSourceWith(user2.IDEQ(settler.ID)),
						transaction.HasDestinationWith(user2.IDEQ(targetUser.ID)),
					),
					transaction.And(
						transaction.HasDestinationWith(user2.IDEQ(settler.ID)),
						transaction.HasSourceWith(user2.IDEQ(targetUser.ID)),
					),
				),
				transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
			).Exec(r.ctx)
	if err != nil {
		return nil, err
	}
	netAmount := existingLentTxn - existingOwedTxn
	if netAmount < 0 {
		netAmount = 0 - netAmount
	}
	if netAmount == 0 {
		err = errors.New("user does not have any outstanding settlements")
		return nil, err
	}
	txnHistory, err := r.app.EntClient.TxnHistory.Create().
		SetAmount(netAmount).
		SetSource(targetUser).
		SetDestination(settler).
		SetBelongsTo(groupObj).
		SetSettled(true).
		SetNote("Settled").
		Save(r.ctx)
	if err != nil {
		return nil, err
	}
	return txnHistory, nil
}
