package db

import (
	"context"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"settlesphere/ent"
	"settlesphere/ent/migrate"
)

func SetUpEnt() *ent.Client {
	dbHost := "localhost"
	dbPort := "5431"
	dbName := "settlesphere-db"
	dbUser := "postgres"
	dbPass := "postgres"
	connString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", dbHost, dbPort, dbName, dbUser, dbPass)
	entClient, err := ent.Open("postgres", connString)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	// Run the auto migration tool.
	if err := entClient.Schema.Create(context.Background(), migrate.WithDropColumn(true), migrate.WithDropIndex(true)); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	return entClient
}
