package services

import (
	"context"
	"errors"
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
