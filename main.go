package main

import (
	"log"
	"os"

	"example/hello/graph/postgres"

	"github.com/gofiber/fiber"
)

func main() {
	postgres.InitDB()
	log.Println("Database initialized successfully")
	app := fiber.New()

	app.Listen(os.Getenv("Port"))
}
