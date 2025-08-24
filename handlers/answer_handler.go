package handlers

import (
	"msu-forum/database"
	"msu-forum/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Criar nova resposta
func CreateAnswer(c *fiber.Ctx) error {
	var data struct {
		Body string `json:"body" validate:"required,min=10"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := Validate.Struct(data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
	}

	questionID, err := strconv.ParseUint(c.Params("questionId"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID da pergunta inválido"})
	}

	userID := c.Locals("user_id").(int)
	now := time.Now()

	// Verificar se a pergunta existe
	var question models.Question
	err = database.DB.Get(&question, "SELECT id FROM questions WHERE id = $1", questionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pergunta não encontrada"})
	}

	// Inserir resposta
	query := `INSERT INTO answers (question_id, user_id, body, votes, is_accepted, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	var answerID uint64
	err = database.DB.QueryRow(
		query,
		questionID, userID, data.Body, 0, false, now, now,
	).Scan(&answerID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar resposta"})
	}

	// Atualizar contador de respostas da pergunta
	database.DB.Exec("UPDATE questions SET answer_count = answer_count + 1 WHERE id = $1", questionID)

	return c.Status(201).JSON(fiber.Map{"id": answerID, "message": "Resposta criada com sucesso"})
}

// Listar respostas de uma pergunta
func GetAnswers(c *fiber.Ctx) error {
	questionID, err := strconv.ParseUint(c.Params("questionId"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID da pergunta inválido"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	query := `
		SELECT a.*, u.username, u.avatar_url
		FROM answers a
		LEFT JOIN users u ON a.user_id = u.id
		WHERE a.question_id = $1
		ORDER BY a.is_accepted DESC, a.votes DESC, a.created_at ASC
		LIMIT $2 OFFSET $3
	`

	var answers []struct {
		models.Answer
		Username  string `json:"username" db:"username"`
		AvatarURL string `json:"avatar_url" db:"avatar_url"`
	}

	err = database.DB.Select(&answers, query, questionID, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar respostas"})
	}

	return c.JSON(answers)
}

// Atualizar resposta
func UpdateAnswer(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}

	userID := c.Locals("user_id").(int)

	var data struct {
		Body string `json:"body" validate:"required,min=10"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := Validate.Struct(data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
	}

	// Verificar se a resposta pertence ao usuário
	var answer models.Answer
	err = database.DB.Get(&answer, "SELECT user_id FROM answers WHERE id = $1", id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Resposta não encontrada"})
	}

	if uint64(userID) != answer.UserID {
		return c.Status(403).JSON(fiber.Map{"error": "Sem permissão para editar esta resposta"})
	}

	// Atualizar resposta
	_, err = database.DB.Exec(
		"UPDATE answers SET body = $1, updated_at = $2 WHERE id = $3",
		data.Body, time.Now(), id,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar resposta"})
	}

	return c.JSON(fiber.Map{"message": "Resposta atualizada com sucesso"})
}

// Deletar resposta
func DeleteAnswer(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}

	userID := c.Locals("user_id").(int)
	role := c.Locals("role").(string)

	// Verificar se a resposta pertence ao usuário ou se é admin
	var answer models.Answer
	err = database.DB.Get(&answer, "SELECT user_id, question_id FROM answers WHERE id = $1", id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Resposta não encontrada"})
	}

	if uint64(userID) != answer.UserID && role != "Admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Sem permissão para deletar esta resposta"})
	}

	// Deletar resposta
	_, err = database.DB.Exec("DELETE FROM answers WHERE id = $1", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao deletar resposta"})
	}

	// Atualizar contador de respostas da pergunta
	database.DB.Exec("UPDATE questions SET answer_count = answer_count - 1 WHERE id = $1", answer.QuestionID)

	return c.JSON(fiber.Map{"message": "Resposta deletada com sucesso"})
}

// Aceitar resposta
func AcceptAnswer(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}

	userID := c.Locals("user_id").(int)

	// Verificar se a resposta existe e buscar a pergunta
	var answer models.Answer
	err = database.DB.Get(&answer, "SELECT question_id FROM answers WHERE id = $1", id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Resposta não encontrada"})
	}

	// Verificar se o usuário é o dono da pergunta
	var question models.Question
	err = database.DB.Get(&question, "SELECT user_id FROM questions WHERE id = $1", answer.QuestionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pergunta não encontrada"})
	}

	if uint64(userID) != question.UserID {
		return c.Status(403).JSON(fiber.Map{"error": "Apenas o autor da pergunta pode aceitar respostas"})
	}

	// Desmarcar outras respostas como aceitas
	database.DB.Exec("UPDATE answers SET is_accepted = false WHERE question_id = $1", answer.QuestionID)

	// Marcar esta resposta como aceita
	_, err = database.DB.Exec("UPDATE answers SET is_accepted = true WHERE id = $1", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao aceitar resposta"})
	}

	// Marcar pergunta como resolvida
	database.DB.Exec("UPDATE questions SET is_solved = true WHERE id = $1", answer.QuestionID)

	return c.JSON(fiber.Map{"message": "Resposta aceita com sucesso"})
}
