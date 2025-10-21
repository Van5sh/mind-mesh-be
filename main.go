package main

import (
	"log"

	"example/hello/graph/postgres"
)

func main() {
	postgres.InitDB()
	log.Println("Database initialized successfully")
}