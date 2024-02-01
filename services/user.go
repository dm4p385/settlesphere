package services

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/ed25519"
	"math/rand"
	"settlesphere/config"
	"settlesphere/ent"
	"settlesphere/ent/group"
	"settlesphere/ent/stat"
	"settlesphere/ent/transaction"
	user2 "settlesphere/ent/user"
	"time"
)

type UserOps struct {
	ctx context.Context
	app *config.Application
}

var defaultProfilePictures = []string{
	"https://cdn.discordapp.com/attachments/876848373720842260/1200728028788031548/7705305_1.png?ex=65c73c1e&is=65b4c71e&hm=eeabe02c18e2acdb2c6d630b5ed3749cc95d69ba9567e406b3e904204ce0c670&",
	"https://cdn.discordapp.com/attachments/876848373720842260/1200728029232644146/7705307_1.png?ex=65c73c1e&is=65b4c71e&hm=1c8617c528e6a6447b07574875fdcc08dbf2a249be105003841ee308bb23d5ee&",
	"https://cdn.discordapp.com/attachments/876848373720842260/1200728029543026759/7705314_1.png?ex=65c73c1e&is=65b4c71e&hm=c13680a0a4cad782610a0d301eda2879f606da8b3980274825a54f8ec30b2019&",
	"https://cdn.discordapp.com/attachments/876848373720842260/1200728029870161991/7705323_1.png?ex=65c73c1e&is=65b4c71e&hm=053e4b8a65c150b8164c88eb4b8e83f25b89a97bb886a258032ab3925ccb7b00&",
	"https://cdn.discordapp.com/attachments/876848373720842260/1200728030121836603/7705329_1.png?ex=65c73c1e&is=65b4c71e&hm=21584b96269c804d8a3bd3afab6f946cddee6556fcdbe4425c01bd62f4fdd68b&",
	"https://cdn.discordapp.com/attachments/876848373720842260/1200728030390276116/7742218_1.png?ex=65c73c1e&is=65b4c71e&hm=6a2cc142ceef64afa654d7f06f48f7ddb45264a4b40be08e97544c2a86230908&",
	"https://cdn.discordapp.com/attachments/876848373720842260/1200728030784536586/7742242_1.png?ex=65c73c1e&is=65b4c71e&hm=650cf1fe297c64878ac46d4be7d0c836471fb8d851e6d28a52089a322d23396f&",
	"https://cdn.discordapp.com/attachments/876848373720842260/1200728031103307916/7748166_1.png?ex=65c73c1f&is=65b4c71f&hm=10617daf678840ed2574fa2d567a35aa6216a7fa6d35da77ff4ae584e8e241e5&",
	"https://cdn.discordapp.com/attachments/876848373720842260/1200728031451414560/7748169_1.png?ex=65c73c1f&is=65b4c71f&hm=5653d203db5ac73730d8e2149dcd4bfc7c61bc90a1a9194cc8e07da3e86848de&",
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

func (r *UserOps) VerifyUser(message string, signatureBase64 string, publicKeyBase64 string) bool {
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		log.Errorf("error occurred while decoding signature: %v", err)
		return false
	}
	//rawPublicKey, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	//if err != nil {
	//	log.Errorf("error occurred while decoding pubKey: %v", err)
	//	return false
	//}
	//signature := base58.Decode(signatureBase64)
	rawPublicKey := base58.Decode(publicKeyBase64)
	log.Debug(signature)
	log.Debug(rawPublicKey)
	// Convert the message to bytes
	messageBytes := []byte(message)
	log.Debug(message)
	log.Debug(messageBytes)
	var publicKey ed25519.PublicKey = rawPublicKey
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
			SetSettledAt(time.Now()).
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
		SetSettledAt(time.Now()).
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

func (r *UserOps) GetProfilePictureUrl() string {
	randomIndex := rand.Intn(len(defaultProfilePictures))
	randomElement := defaultProfilePictures[randomIndex]
	return randomElement
}

func (r *UserOps) UpdateUserPaidByStat(totalAmount float64, userObj *ent.User, groupObj *ent.Group) error {
	stat, err := r.app.EntClient.Stat.Query().Where(
		stat.HasBelongsToGroupWith(group.IDEQ(groupObj.ID)),
		stat.HasBelongsToUserWith(user2.IDEQ(userObj.ID)),
	).Only(r.ctx)
	if err != nil {
		return err
	}
	log.Debug(stat)
	_, err = stat.Update().AddTotalPaid(totalAmount).Save(r.ctx)
	return nil
}

func (r *UserOps) UpdateUserShareStat(userShare float64, userObj *ent.User, groupObj *ent.Group) error {
	stat, err := r.app.EntClient.Stat.Query().Where(
		stat.HasBelongsToGroupWith(group.IDEQ(groupObj.ID)),
		stat.HasBelongsToUserWith(user2.IDEQ(userObj.ID)),
	).Only(r.ctx)
	if err != nil {
		return err
	}
	log.Debug(stat)
	_, err = stat.Update().AddTotalShare(userShare).Save(r.ctx)
	return nil
}
