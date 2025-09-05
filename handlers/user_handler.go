package handlers

import (
	"msu-forum/database"
	"msu-forum/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Obter perfil do usuário atual
func GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	var user models.User
	err := database.DB.Get(&user, "SELECT * FROM users WHERE id = $1", userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Usuário não encontrado"})
	}

	return c.JSON(user)
}

func HasUserWithThisWallet(c *fiber.Ctx) error {
	var body struct {
		Wallet string `json:"wallet"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	var exists bool
	err := database.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE wallet = $1)", body.Wallet)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao consultar banco"})
	}

	return c.JSON(fiber.Map{"exists": exists})
}

// Atualizar perfil do usuário
func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	var data struct {
		Username  string `json:"username"`
		Phone     string `json:"phone"`
		Wallet    string `json:"wallet"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	// Atualizar perfil
	query := `UPDATE users SET username = $1, phone = $2, wallet = $3, avatar_url = $4, last_seen = $5 
			  WHERE id = $6`

	_, err := database.DB.Exec(query, data.Username, data.Phone, data.Wallet, data.AvatarURL, time.Now(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar perfil"})
	}

	return c.JSON(fiber.Map{"message": "Perfil atualizado com sucesso"})
}

// Obter perguntas de um usuário
func GetUserQuestions(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("userId"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID do usuário inválido"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	query := `
		SELECT q.*, u.username, u.avatar_url,
		       COUNT(DISTINCT a.id) as answer_count
		FROM questions q
		LEFT JOIN users u ON q.user_id = u.id
		LEFT JOIN answers a ON q.id = a.question_id
		WHERE q.user_id = $1
		GROUP BY q.id, u.username, u.avatar_url
		ORDER BY q.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var questions []struct {
		models.Question
		Username  string `json:"username" db:"username"`
		AvatarURL string `json:"avatar_url" db:"avatar_url"`
	}

	err = database.DB.Select(&questions, query, userID, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar perguntas"})
	}

	return c.JSON(questions)
}

// Obter respostas de um usuário
func GetUserAnswers(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("userId"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID do usuário inválido"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	query := `
		SELECT a.*, u.username, u.avatar_url, q.title as question_title
		FROM answers a
		LEFT JOIN users u ON a.user_id = u.id
		LEFT JOIN questions q ON a.question_id = q.id
		WHERE a.user_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var answers []struct {
		models.Answer
		Username      string `json:"username" db:"username"`
		AvatarURL     string `json:"avatar_url" db:"avatar_url"`
		QuestionTitle string `json:"question_title" db:"question_title"`
	}

	err = database.DB.Select(&answers, query, userID, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar respostas"})
	}

	return c.JSON(answers)
}

// Listar usuários (apenas admin)
func GetUsers(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "Admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Apenas administradores podem listar usuários"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset := (page - 1) * limit

	query := `
		SELECT id, username, email, reputation, role, phone, wallet, created_at, last_seen, is_active, avatar_url
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var users []models.User
	err := database.DB.Select(&users, query, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar usuários"})
	}

	return c.JSON(users)
}

// Atualizar status do usuário (apenas admin)
func UpdateUserStatus(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "Admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Apenas administradores podem atualizar status de usuários"})
	}

	userID, err := strconv.ParseUint(c.Params("userId"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID do usuário inválido"})
	}

	var data struct {
		IsActive bool   `json:"is_active"`
		Role     string `json:"role"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	// Validar role
	validRoles := map[string]bool{
		"Admin": true, "Streamer": true, "Moderator": true, "Member": true,
	}
	if !validRoles[data.Role] {
		return c.Status(400).JSON(fiber.Map{"error": "Role inválido"})
	}

	// Atualizar status
	_, err = database.DB.Exec(
		"UPDATE users SET is_active = $1, role = $2 WHERE id = $3",
		data.IsActive, data.Role, userID,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar status do usuário"})
	}

	return c.JSON(fiber.Map{"message": "Status do usuário atualizado com sucesso"})
}
