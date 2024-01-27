package config

import (
	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2"
	"settlesphere/ent"
)

type Db struct {
	DbHost string
	DbPort string
	DbName string
	DbUser string
	DbPass string
}

type Application struct {
	Db        Db
	FiberApp  *fiber.App
	EntClient *ent.Client
	//FirebaseApp           *firebase.App
	FirebaseStorageClient *storage.Client
	//log       *log.Logger
	Secret string
}

func InitializeApp(fiberApp *fiber.App, entClient *ent.Client, firebaseStorageClient *storage.Client) *Application {
	app := Application{
		Db: Db{
			DbHost: "localhost",
			DbPort: "5438",
			DbName: "settlesphere-db",
			DbUser: "postgres",
			DbPass: "postgres",
		},
		FiberApp:  fiberApp,
		EntClient: entClient,
		//log:       log,
		Secret: "GODLESSPLANET",
		//FirebaseApp:           firebaseApp,
		FirebaseStorageClient: firebaseStorageClient,
	}
	return &app
}
