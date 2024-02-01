package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
	"os"
	"settlesphere/config"
	"settlesphere/db"
	"settlesphere/routers"
	"settlesphere/services"
)

//func loadEnv() error {
//	err := godotenv.Load(".env")
//	if err != nil {
//		return fmt.Errorf("Error loading .env file: %v", err)
//	}
//	return nil
//}

func main() {
	envType := os.Getenv("ENV_TYPE")
	log.Debugf("environment type: %s", envType)

	entClient := db.SetUpEnt()
	defer entClient.Close()
	fiberApp := fiber.New(fiber.Config{
		ServerHeader: "SettleSphere",
		AppName:      "SettleSphere",
	})
	fiberApp.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin,Authorization",
		AllowOrigins:     "*",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	//firebaseApp, err := services.InitFirebase()
	//if err != nil {
	//	panic(err)
	//}
	firebaseStorageClient, err := services.InitStorageClient()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	app := config.InitializeApp(fiberApp, entClient, firebaseStorageClient)
	fiberApp.Use(recover2.New())
	// setup routes
	routers.SetRoutes(app)
	fiberApp.Listen(":3000")
}
