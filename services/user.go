package services

import (
	"context"
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/ed25519"
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
	userObj, err := r.app.EntClient.User.Query().Where(user2.PubKeyEQ(claims["pubkey"].(string))).Only(ctx)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return userObj, nil
}

func (r *UserOps) VerifyUser(message string, signatureBase58 string, publicKeyBase58 string) bool {
	signature := base58.Decode(signatureBase58)
	publicKey := base58.Decode(publicKeyBase58)
	log.Debug(signature)
	log.Debug(publicKey)
	// Convert the message to bytes
	messageBytes := []byte(message)
	log.Debug(message)
	log.Debug(messageBytes)
	// Perform signature verification
	verified := ed25519.Verify(publicKey, messageBytes, signature)
	if verified {
		log.Debug("Signature is valid.")
		return true
	} else {
		log.Debug("Signature is not valid.")
		return false
	}
}

type txn struct {
	Owes     []*ent.Transaction `json:"owes"`
	Receives []*ent.Transaction `json:"receives"`
}

type userInfoTxn struct {
	SourceUserId      int    `json:"source_user_id"`
	SourcePubKey      string `json:"source_pub_key"`
	DestinationUserId int    `json:"destination_user_id"`
	DestinationPubKey string `json:"destination_pub_key"`
}

func (r *UserOps) GetUserTxns(user *ent.User, groupObj *ent.Group) (txn, error) {
	lentTxns, err := r.app.EntClient.Transaction.Query().
		Where(
			transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
			transaction.HasDestinationWith(user2.IDEQ(user.ID)),
		).WithDestination().WithSource().All(r.ctx)
	if err != nil {
		return txn{}, err
	}
	owedTxns, err := r.app.EntClient.Transaction.Query().
		Where(
			transaction.HasBelongsToWith(group.IDEQ(groupObj.ID)),
			transaction.HasSourceWith(user2.IDEQ(user.ID)),
		).WithDestination().WithSource().All(r.ctx)
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
func (r *UserOps) SettleTxn(settler *ent.User, targetUser *ent.User, groupObj *ent.Group) (*ent.TxnHistory, *userInfoTxn, error) {
	existingLentTxn := 0.0
	existingOwedTxn := 0.0
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
			).Aggregate(ent.Sum(transaction.FieldAmount)).Float64(r.ctx)
		if err != nil {
			log.Error(err)
			return nil, nil, err
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
			).Aggregate(ent.Sum(transaction.FieldAmount)).Float64(r.ctx)
		if err != nil {
			return nil, nil, err
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
		return nil, nil, err
	}
	netAmount := existingLentTxn - existingOwedTxn
	if netAmount < 0 {
		netAmount = 0 - netAmount
		txnHistory, err := r.app.EntClient.TxnHistory.Create().
			SetAmount(netAmount).
			SetSource(settler).
			SetDestination(targetUser).
			SetBelongsTo(groupObj).
			SetSettled(true).
			SetNote("Settled").
			Save(r.ctx)
		if err != nil {
			return nil, nil, err
		}
		usrInfoTxn := userInfoTxn{
			SourceUserId:      settler.ID,
			SourcePubKey:      settler.PubKey,
			DestinationUserId: targetUser.ID,
			DestinationPubKey: targetUser.PubKey,
		}
		return txnHistory, &usrInfoTxn, nil
	}

	if netAmount == 0 {
		err = errors.New("user does not have any outstanding settlements")
		return nil, nil, err
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
		return nil, nil, err
	}
	usrInfoTxn := userInfoTxn{
		SourceUserId:      targetUser.ID,
		SourcePubKey:      targetUser.PubKey,
		DestinationUserId: settler.ID,
		DestinationPubKey: settler.PubKey,
	}

	return txnHistory, &usrInfoTxn, nil
}
