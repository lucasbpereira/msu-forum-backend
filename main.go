package main

import (
	"log"
	"msu-forum/database"
	"msu-forum/handlers"
	"msu-forum/middleware"
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

	// Middlewares
	app.Use(middleware.CORSMiddleware())
	app.Static("/assets", "./assets")
	// Rotas públicas
	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)
	app.Get("/questions", handlers.GetQuestions)
	app.Get("/questions/:id", handlers.GetQuestion)
	app.Get("/tags", handlers.GetTags)
	app.Get("/tags/:id", handlers.GetTag)
	app.Get("/tags/:tagId/questions", handlers.GetQuestionsByTag)

	// Rotas protegidas
	api := app.Group("/api", middleware.AuthRequired)
	v1 := api.Group("/v1")
	// Perguntas
	v1.Post("/questions", handlers.CreateQuestion)
	v1.Put("/questions/:id", handlers.UpdateQuestion)
	v1.Delete("/questions/:id", handlers.DeleteQuestion)

	// Respostas
	v1.Post("/questions/:questionId/answers", handlers.CreateAnswer)
	v1.Get("/questions/:questionId/answers", handlers.GetAnswers)
	v1.Put("/answers/:id", handlers.UpdateAnswer)
	v1.Delete("/answers/:id", handlers.DeleteAnswer)
	v1.Post("/answers/:id/accept", handlers.AcceptAnswer)

	// Votos
	v1.Post("/votes", handlers.Vote)
	v1.Get("/votes", handlers.GetUserVotes)

	// Usuários
	v1.Get("/profile", handlers.GetProfile)
	v1.Put("/profile", handlers.UpdateProfile)
	v1.Get("/users/:userId/questions", handlers.GetUserQuestions)
	v1.Get("/users/:userId/answers", handlers.GetUserAnswers)

	// Admin routes
	admin := v1.Group("/admin", func(c *fiber.Ctx) error {
		role := c.Locals("role").(string)
		if role != "Admin" {
			return c.Status(403).JSON(fiber.Map{"error": "Acesso negado"})
		}
		return c.Next()
	})

	admin.Get("/users", handlers.GetUsers)
	admin.Put("/users/:userId/status", handlers.UpdateUserStatus)
	admin.Post("/tags", handlers.CreateTag)
	admin.Put("/tags/:id", handlers.UpdateTag)
	admin.Delete("/tags/:id", handlers.DeleteTag)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000" // porta padrão
	}

	log.Printf("🚀 Servidor rodando na porta %s", port)
	app.Listen(":" + port)
}
