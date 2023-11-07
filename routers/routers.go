package routers

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"settlesphere/config"
	"settlesphere/handlers"
)

func SetRoutes(app *config.Application) {
	api := app.FiberApp.Group("/api/v1/")
	jwtMiddleware := jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(app.Secret)},
	})
	// server routes
	api.Get("/status", handlers.Status)

	// auth routes
	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login(app))

	// group routes
	group := api.Group("/groups")
	group.Use(jwtMiddleware)
	group.Get("/", handlers.ListGroups(app))
	group.Get("/join/:code", handlers.JoinGroup(app))
	group.Post("/", handlers.CreateGroup(app))
	group.Get("/members/:code", handlers.GetUsers(app))

	// transaction routes
	txn := api.Group("/txn")
	txn.Use(jwtMiddleware)
	//txn.Get("/", handlers.ListTxns(app))
	txn.Get("/group/:code", handlers.GroupUserTxns(app))
	txn.Post("/group/:code", handlers.AddTransaction(app))
	txn.Get("/group/:code/history", handlers.TxnHistory(app))
	//txn.Get("/group/:code/settle", handlers.)
}
