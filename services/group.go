package services

import (
	"context"
	"settlesphere/config"
	"settlesphere/ent"
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

//func (r *GroupOps) GetAllGroupTxns(group *ent.Group) ([]ent.Transaction, error) {
//	txns, err := group.QueryTransactions().Select().All(r.ctx)
//	if err:=
//}
