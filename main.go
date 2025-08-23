package main

import (
	"log"
	"msu-forum/database"
	"msu-forum/handlers"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar arquivo .env")
	}

	database.Connect()

	app := fiber.New()

	// Auth
	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)

	// Rota protegida
	// app.Get("/me", middleware.AuthRequired, handlers.GetMe)
	// app.Get("/me/members", middleware.AuthRequired, handlers.GetMembers)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000" // porta padrão
	}

	app.Listen(":" + port)
}
