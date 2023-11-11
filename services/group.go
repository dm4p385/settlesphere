package services

import (
	"context"
	"settlesphere/config"
	"settlesphere/ent"
	"settlesphere/ent/txnhistory"
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
	OwedTxnHistory, err := user.QueryOwedHistory().Where(txnhistory.Settled(true)).All(r.ctx)
	if err != nil {
		return nil, nil, err
	}
	LentTxnHistory, err := user.QueryLentHistory().Where(txnhistory.Settled(true)).All(r.ctx)
	if err != nil {
		return nil, nil, err
	}
	return OwedTxnHistory, LentTxnHistory, nil
}

//func (r *GroupOps) GetAllGroupTxns(group *ent.Group) ([]ent.Transaction, error) {
//	txns, err := group.QueryTransactions().Select().All(r.ctx)
//	if err:=
//}
