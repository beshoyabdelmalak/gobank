package main

import (
	"log"
)

func main() {
	db, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	apiServer := NewAPIServer(":8000", db)
	apiServer.Run()
}
