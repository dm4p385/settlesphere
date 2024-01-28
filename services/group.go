package services

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"settlesphere/config"
	"settlesphere/ent"
	"settlesphere/ent/group"
	"settlesphere/ent/transaction"
	"settlesphere/ent/txnhistory"
	user2 "settlesphere/ent/user"
)

type GroupOps struct {
	ctx context.Context
	app *config.Application
}

func NewGroupOps(ctx context.Context, app *config.Application) *GroupOps {
	return &GroupOps{
		ctx: ctx,
		app: app,
	}
}

func (r *GroupOps) AddUserToGroup(group *ent.Group, user *ent.User) error {
	_, err := group.Update().AddUsers(user).Save(r.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *GroupOps) GetTxnHistoryOfGroup(group *ent.Group) ([]*ent.TxnHistory, error) {
	txnHistory, err := group.QueryTxnHistory().All(r.ctx)
	if err != nil {
		return nil, err
	}
	return txnHistory, nil
}

func (r *GroupOps) GetSettledTxnsOfAllGroups(user *ent.User) ([]*ent.TxnHistory, []*ent.TxnHistory, error) {
	OwedTxnHistory, err := user.QueryOwedHistory().Where(txnhistory.Settled(true)).WithBelongsTo().All(r.ctx)
	if err != nil {
		return nil, nil, err
	}
	LentTxnHistory, err := user.QueryLentHistory().Where(txnhistory.Settled(true)).WithBelongsTo().All(r.ctx)
	if err != nil {
		return nil, nil, err
	}
	return OwedTxnHistory, LentTxnHistory, nil
}

func (r *GroupOps) GetUserNetAmountOfGroup(user *ent.User, groupObj *ent.Group) (float64, error) {
	owedTxnsAmount, err := r.app.EntClient.Transaction.Query().
		Where(
			transaction.HasDestinationWith(user2.ID(user.ID)),
			transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
		).
		Aggregate(ent.Sum(transaction.FieldAmount)).Float64(r.ctx)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	lentTxnsAmount, err := r.app.EntClient.Transaction.Query().
		Where(
			transaction.HasSourceWith(user2.ID(user.ID)),
			transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
		).
		Aggregate(ent.Sum(transaction.FieldAmount)).Float64(r.ctx)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	netAmount := lentTxnsAmount - owedTxnsAmount
	return netAmount, nil
}

//func (r *GroupOps) GetAllGroupTxns(group *ent.Group) ([]ent.Transaction, error) {
//	txns, err := group.QueryTransactions().Select().All(r.ctx)
//	if err:=
//}
