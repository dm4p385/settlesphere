package services

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"settlesphere/config"
	"settlesphere/ent"
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
