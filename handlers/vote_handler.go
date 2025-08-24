package handlers

import (
	"msu-forum/database"
	"msu-forum/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Votar em uma pergunta ou resposta
func Vote(c *fiber.Ctx) error {
	var data struct {
		PostType string `json:"post_type"` // "question" ou "answer"
		PostID   uint64 `json:"post_id"`
		Type     int8   `json:"type"` // +1 ou -1
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	// Validações
	if data.PostType != "question" && data.PostType != "answer" {
		return c.Status(400).JSON(fiber.Map{"error": "post_type deve ser 'question' ou 'answer'"})
	}

	if data.Type != 1 && data.Type != -1 {
		return c.Status(400).JSON(fiber.Map{"error": "type deve ser 1 (upvote) ou -1 (downvote)"})
	}

	userID := c.Locals("user_id").(int)
	now := time.Now()

	// Verificar se já existe um voto do usuário
	var existingVote models.Vote
	err := database.DB.Get(&existingVote,
		"SELECT * FROM votes WHERE user_id = $1 AND post_id = $2 AND post_type = $3",
		userID, data.PostID, data.PostType)

	if err == nil {
		// Voto já existe, atualizar
		if existingVote.Type == data.Type {
			// Mesmo tipo de voto, remover o voto
			database.DB.Exec("DELETE FROM votes WHERE id = $1", existingVote.ID)

			// Atualizar contador de votos
			if data.PostType == "question" {
				database.DB.Exec("UPDATE questions SET votes = votes - $1 WHERE id = $2", existingVote.Type, data.PostID)
			} else {
				database.DB.Exec("UPDATE answers SET votes = votes - $1 WHERE id = $2", existingVote.Type, data.PostID)
			}

			return c.JSON(fiber.Map{"message": "Voto removido"})
		} else {
			// Tipo diferente, atualizar voto
			database.DB.Exec("UPDATE votes SET type = $1, created_at = $2 WHERE id = $3",
				data.Type, now, existingVote.ID)

			// Atualizar contador de votos (diferença de 2)
			voteDiff := data.Type - existingVote.Type
			if data.PostType == "question" {
				database.DB.Exec("UPDATE questions SET votes = votes + $1 WHERE id = $2", voteDiff, data.PostID)
			} else {
				database.DB.Exec("UPDATE answers SET votes = votes + $1 WHERE id = $2", voteDiff, data.PostID)
			}

			return c.JSON(fiber.Map{"message": "Voto atualizado"})
		}
	}

	// Criar novo voto
	_, err = database.DB.Exec(
		"INSERT INTO votes (user_id, post_id, post_type, type, created_at) VALUES ($1, $2, $3, $4, $5)",
		userID, data.PostID, data.PostType, data.Type, now,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar voto"})
	}

	// Atualizar contador de votos
	if data.PostType == "question" {
		database.DB.Exec("UPDATE questions SET votes = votes + $1 WHERE id = $2", data.Type, data.PostID)
	} else {
		database.DB.Exec("UPDATE answers SET votes = votes + $1 WHERE id = $2", data.Type, data.PostID)
	}

	return c.Status(201).JSON(fiber.Map{"message": "Voto registrado com sucesso"})
}

// Obter votos de um usuário
func GetUserVotes(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	query := `
		SELECT v.*, 
		       CASE 
		           WHEN v.post_type = 'question' THEN q.title
		           WHEN v.post_type = 'answer' THEN a.body
		       END as post_content
		FROM votes v
		LEFT JOIN questions q ON v.post_type = 'question' AND v.post_id = q.id
		LEFT JOIN answers a ON v.post_type = 'answer' AND v.post_id = a.id
		WHERE v.user_id = $1
		ORDER BY v.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var votes []struct {
		models.Vote
		PostContent string `json:"post_content" db:"post_content"`
	}

	err := database.DB.Select(&votes, query, userID, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar votos"})
	}

	return c.JSON(votes)
}
