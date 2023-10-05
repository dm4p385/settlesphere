package db

import (
	"context"
	"log"
	"settlesphere/ent"
)

func SetUpEnt() {
	entClient, err := ent.Open("postgres", "host=settlesphere-db port=5432 dbname=settlesphere-db password=postgres")
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer entClient.Close()
	// Run the auto migration tool.
	if err := entClient.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
}
