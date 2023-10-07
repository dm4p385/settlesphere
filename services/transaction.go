package services

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"settlesphere/config"
	"settlesphere/ent"
	"settlesphere/ent/group"
	"settlesphere/ent/transaction"
	user2 "settlesphere/ent/user"
)

type TxnOps struct {
	ctx context.Context
	app *config.Application
}

func NewTxnOps(ctx context.Context, app *config.Application) *TxnOps {
	return &TxnOps{
		ctx: ctx,
		app: app,
	}
}

func (r *TxnOps) GenerateTransaction(groupObj *ent.Group, sourceUser *ent.User, destUser *ent.User, amount int) (*ent.Transaction, error) {
	existingLentTxn := 0
	existingOwedTxn := 0
	var err error
	if temp := r.app.EntClient.Transaction.Query().
		Where(
			transaction.And(
				transaction.HasSourceWith(user2.IDEQ(sourceUser.ID)),
				transaction.HasDestinationWith(user2.IDEQ(destUser.ID)),
			),
			transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
		).ExistX(r.ctx); temp {
		existingLentTxn, err = r.app.EntClient.Transaction.Query().
			Where(
				transaction.And(
					transaction.HasSourceWith(user2.IDEQ(sourceUser.ID)),
					transaction.HasDestinationWith(user2.IDEQ(destUser.ID)),
				),
				transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
			).Aggregate(ent.Sum(transaction.FieldAmount)).Int(r.ctx)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	//existingLentTxn, err := r.app.EntClient.Transaction.Query().
	//	Where(
	//		transaction.And(
	//			transaction.HasSourceWith(user2.IDEQ(sourceUser.ID)),
	//			transaction.HasDestinationWith(user2.IDEQ(destUser.ID)),
	//		),
	//	).Aggregate(ent.Sum(transaction.FieldAmount)).Int(r.ctx)
	//if err != nil {
	//	log.Error(err)
	//	return nil, err
	//}
	if temp := r.app.EntClient.Transaction.Query().
		Where(
			transaction.And(
				transaction.HasDestinationWith(user2.IDEQ(sourceUser.ID)),
				transaction.HasSourceWith(user2.IDEQ(destUser.ID)),
			),
			transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
		).ExistX(r.ctx); temp {
		existingOwedTxn, err = r.app.EntClient.Transaction.Query().
			Where(
				transaction.And(
					transaction.HasDestinationWith(user2.IDEQ(sourceUser.ID)),
					transaction.HasSourceWith(user2.IDEQ(destUser.ID)),
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
						transaction.HasSourceWith(user2.IDEQ(sourceUser.ID)),
						transaction.HasDestinationWith(user2.IDEQ(destUser.ID)),
					),
					transaction.And(
						transaction.HasDestinationWith(user2.IDEQ(sourceUser.ID)),
						transaction.HasSourceWith(user2.IDEQ(destUser.ID)),
					),
				),
				transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
			).Exec(r.ctx)
	if err != nil {
		return nil, err
	}
	netAmount := existingLentTxn - existingOwedTxn + amount
	log.Debug(netAmount, existingLentTxn, existingOwedTxn, amount)
	if netAmount > 0 {
		txn, err := r.app.EntClient.Transaction.Create().
			SetAmount(netAmount).
			AddSource(sourceUser).
			AddDestination(destUser).
			AddBelongsTo(groupObj).
			Save(r.ctx)
		if err != nil {
			return nil, err
		}
		return txn, nil
	} else if netAmount < 0 {
		txn, err := r.app.EntClient.Transaction.Create().
			SetAmount(0 - netAmount).
			AddSource(destUser).
			AddDestination(sourceUser).
			AddBelongsTo(groupObj).
			Save(r.ctx)
		if err != nil {
			return nil, err
		}
		return txn, nil
	} else if netAmount == 0 {
		return nil, nil
	}
	return nil, nil
}
