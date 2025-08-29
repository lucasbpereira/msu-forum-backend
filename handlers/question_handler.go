package handlers

import (
	"fmt"
	"msu-forum/database"
	"msu-forum/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Criar nova pergunta
func CreateQuestion(c *fiber.Ctx) error {
	var data struct {
		Title string   `json:"title" validate:"required,min=5,max=200"`
		Body  string   `json:"body" validate:"required,min=10"`
		Tags  []string `json:"tags" validate:"max=5"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := Validate.Struct(data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos", "details": err.Error()})
	}

	userID := c.Locals("user_id").(int)
	now := time.Now()

	// Inserir pergunta
	query := `INSERT INTO questions (user_id, title, body, votes, view_count, answer_count, is_solved, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	var questionID uint64
	err := database.DB.QueryRow(
		query,
		userID, data.Title, data.Body, 0, 0, 0, false, now, now,
	).Scan(&questionID)

	if err != nil {
		fmt.Printf("Erro no banco: %v\n", err)
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar pergunta"})
	}

	// Inserir tags se fornecidas
	if len(data.Tags) > 0 {
		for _, tagName := range data.Tags {
			// Verificar se tag existe, se não, criar
			var tagID uint64
			err := database.DB.QueryRow("SELECT id FROM tags WHERE name = $1", tagName).Scan(&tagID)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Tag não encontrada"})
			} else {
				// Atualizar contador de uso
				database.DB.Exec("UPDATE tags SET usage_count = usage_count + 1 WHERE id = $1", tagID)
			}

			// Inserir relação question_tag
			database.DB.Exec("INSERT INTO question_tags (question_id, tag_id) VALUES ($1, $2)", questionID, tagID)
		}
	}

	return c.Status(201).JSON(fiber.Map{"id": questionID, "message": "Pergunta criada com sucesso"})
}

// Listar perguntas
func GetQuestions(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	fmt.Printf("Executando query com limit: %d, offset: %d\n", limit, offset)
	query := `
		SELECT q.*, u.username, u.avatar_url,
		       COUNT(DISTINCT a.id) as answer_count
		FROM questions q
		LEFT JOIN users u ON q.user_id = u.id
		LEFT JOIN answers a ON q.id = a.question_id
		GROUP BY q.id, u.username, u.avatar_url
		ORDER BY q.created_at DESC
		LIMIT $1 OFFSET $2
	`

	var questions []struct {
		models.Question
		Username  string `json:"username" db:"username"`
		AvatarURL string `json:"avatar_url" db:"avatar_url"`
	}

	err := database.DB.Select(&questions, query, limit, offset)
	if err != nil {
		fmt.Printf("Erro no banco: %v\n", err)
		return c.Status(500).JSON(fiber.Map{
			"error":   "Erro ao buscar perguntas",
			"details": err.Error(), // Mostra o erro real (apenas em desenvolvimento)
		})
	}

	return c.JSON(questions)
}

// Buscar pergunta por ID
func GetQuestion(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}

	// Incrementar contador de visualizações
	database.DB.Exec("UPDATE questions SET view_count = view_count + 1 WHERE id = $1", id)

	// Buscar pergunta com usuário
	var question struct {
		models.Question
		Username  string `json:"username" db:"username"`
		AvatarURL string `json:"avatar_url" db:"avatar_url"`
	}

	err = database.DB.Get(&question, `
		SELECT q.*, u.username, u.avatar_url
		FROM questions q
		LEFT JOIN users u ON q.user_id = u.id
		WHERE q.id = $1
	`, id)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pergunta não encontrada"})
	}

	// Buscar tags da pergunta
	var tags []models.Tag
	database.DB.Select(&tags, `
		SELECT t.*
		FROM tags t
		JOIN question_tags qt ON t.id = qt.tag_id
		WHERE qt.question_id = $1
	`, id)
	question.Tags = tags

	// Buscar respostas
	var answers []models.Answer
	database.DB.Select(&answers, `
		SELECT a.*, u.username, u.avatar_url
		FROM answers a
		LEFT JOIN users u ON a.user_id = u.id
		WHERE a.question_id = $1
		ORDER BY a.is_accepted DESC, a.votes DESC, a.created_at ASC
	`, id)
	question.Answers = answers

	return c.JSON(question)
}

// Atualizar pergunta
func UpdateQuestion(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}

	userID := c.Locals("user_id").(int)

	var data struct {
		Title string   `json:"title" validate:"required,min=5,max=200"`
		Body  string   `json:"body" validate:"required,min=10"`
		Tags  []string `json:"tags" validate:"max=5"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := Validate.Struct(data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
	}

	// Verificar se a pergunta pertence ao usuário
	var question models.Question
	err = database.DB.Get(&question, "SELECT user_id FROM questions WHERE id = $1", id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pergunta não encontrada"})
	}

	if uint64(userID) != question.UserID {
		return c.Status(403).JSON(fiber.Map{"error": "Sem permissão para editar esta pergunta"})
	}

	// Atualizar pergunta
	_, err = database.DB.Exec(
		"UPDATE questions SET title = $1, body = $2, updated_at = $3 WHERE id = $4",
		data.Title, data.Body, time.Now(), id,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar pergunta"})
	}

	return c.JSON(fiber.Map{"message": "Pergunta atualizada com sucesso"})
}

// Deletar pergunta
func DeleteQuestion(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}

	userID := c.Locals("user_id").(int)
	role := c.Locals("role").(string)

	// Verificar se a pergunta pertence ao usuário ou se é admin
	var question models.Question
	err = database.DB.Get(&question, "SELECT user_id FROM questions WHERE id = $1", id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pergunta não encontrada"})
	}

	if uint64(userID) != question.UserID && role != "Admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Sem permissão para deletar esta pergunta"})
	}

	// Deletar pergunta (cascade irá deletar respostas e votos)
	_, err = database.DB.Exec("DELETE FROM questions WHERE id = $1", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao deletar pergunta"})
	}

	return c.JSON(fiber.Map{"message": "Pergunta deletada com sucesso"})
}


// Buscar perguntas por título ou corpo
func SearchQuestions(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Parâmetro de busca é obrigatório"})
	}

	var questions []struct {
		models.Question
		Username  string `json:"username" db:"username"`
		AvatarURL string `json:"avatar_url" db:"avatar_url"`
		Similarity float64 `json:"similarity" db:"similarity"`
	}

	err := database.DB.Select(&questions, `
		SELECT q.*, u.username, u.avatar_url,
		       GREATEST(similarity(q.title, $1), similarity(q.body, $1)) AS similarity
		FROM questions q
		LEFT JOIN users u ON q.user_id = u.id
		WHERE q.title ILIKE '%' || $1 || '%' OR q.body ILIKE '%' || $1 || '%'
		ORDER BY similarity DESC
		LIMIT 20
	`, query)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar perguntas"})
	}

	return c.JSON(questions)
}
