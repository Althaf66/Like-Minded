package main

import (
	"log"

	"github.com/Althaf66/go-backend-social-app/internal/db"
	"github.com/Althaf66/go-backend-social-app/internal/store"
)

func main() {
	conn, err := db.New("postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable", 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewPostgresStorage(conn)

	db.Seed(store, conn)
}
