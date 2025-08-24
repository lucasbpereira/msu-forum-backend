package handlers

import (
	"fmt"
	"msu-forum/database"
	"msu-forum/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Listar todas as tags
func GetTags(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset := (page - 1) * limit

	query := `
		    SELECT t.*, COUNT(qt.question_id) as usage_count
			FROM tags t
			LEFT JOIN question_tags qt ON t.id = qt.tag_id
			GROUP BY t.id
			ORDER BY 2 DESC, t.name ASC  -- 2 refere-se à segunda coluna (usage_count)
			LIMIT $1 OFFSET $2
	`

	var tags []models.Tag
	err := database.DB.Select(&tags, query, limit, offset)
	if err != nil {
		fmt.Printf("Erro no banco: %v\n", err)
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar tags"})
	}

	return c.JSON(tags)
}

// Buscar tag por ID
func GetTag(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}

	var tag models.Tag
	err = database.DB.Get(&tag, "SELECT * FROM tags WHERE id = $1", id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Tag não encontrada"})
	}

	return c.JSON(tag)
}

// Buscar perguntas por tag
func GetQuestionsByTag(c *fiber.Ctx) error {
	tagID, err := strconv.ParseUint(c.Params("tagId"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID da tag inválido"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	query := `
		SELECT q.*, u.username, u.avatar_url,
		       COUNT(DISTINCT a.id) as answer_count
		FROM questions q
		JOIN question_tags qt ON q.id = qt.question_id
		LEFT JOIN users u ON q.user_id = u.id
		LEFT JOIN answers a ON q.id = a.question_id
		WHERE qt.tag_id = $1
		GROUP BY q.id, u.username, u.avatar_url
		ORDER BY q.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var questions []struct {
		models.Question
		Username  string `json:"username" db:"username"`
		AvatarURL string `json:"avatar_url" db:"avatar_url"`
	}

	err = database.DB.Select(&questions, query, tagID, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar perguntas"})
	}

	return c.JSON(questions)
}

// Criar nova tag (apenas admin)
func CreateTag(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "Admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Apenas administradores podem criar tags"})
	}

	var data struct {
		Name        string `json:"name" validate:"required,min=2,max=50"`
		Description string `json:"description" validate:"max=200"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := Validate.Struct(data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
	}

	// Verificar se tag já existe
	var existingTag models.Tag
	err := database.DB.Get(&existingTag, "SELECT id FROM tags WHERE name = $1", data.Name)
	if err == nil {
		return c.Status(409).JSON(fiber.Map{"error": "Tag já existe"})
	}

	// Criar tag
	query := `INSERT INTO tags (name, description, usage_count, created_at) 
			  VALUES ($1, $2, $3, $4) RETURNING id`

	var tagID uint64
	err = database.DB.QueryRow(
		query,
		data.Name, data.Description, 0, time.Now(),
	).Scan(&tagID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar tag"})
	}

	return c.Status(201).JSON(fiber.Map{"id": tagID, "message": "Tag criada com sucesso"})
}

// Atualizar tag (apenas admin)
func UpdateTag(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "Admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Apenas administradores podem editar tags"})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}

	var data struct {
		Name        string `json:"name" validate:"required,min=2,max=50"`
		Description string `json:"description" validate:"max=200"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if err := Validate.Struct(data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
	}

	// Verificar se tag existe
	var tag models.Tag
	err = database.DB.Get(&tag, "SELECT id FROM tags WHERE id = $1", id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Tag não encontrada"})
	}

	// Atualizar tag
	_, err = database.DB.Exec(
		"UPDATE tags SET name = $1, description = $2 WHERE id = $3",
		data.Name, data.Description, id,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar tag"})
	}

	return c.JSON(fiber.Map{"message": "Tag atualizada com sucesso"})
}

// Deletar tag (apenas admin)
func DeleteTag(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "Admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Apenas administradores podem deletar tags"})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
	}

	// Verificar se tag existe
	var tag models.Tag
	err = database.DB.Get(&tag, "SELECT id FROM tags WHERE id = $1", id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Tag não encontrada"})
	}

	// Deletar relações question_tags primeiro
	database.DB.Exec("DELETE FROM question_tags WHERE tag_id = $1", id)

	// Deletar tag
	_, err = database.DB.Exec("DELETE FROM tags WHERE id = $1", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao deletar tag"})
	}

	return c.JSON(fiber.Map{"message": "Tag deletada com sucesso"})
}
