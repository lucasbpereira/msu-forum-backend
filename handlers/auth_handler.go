package handlers

import (
	"msu-forum/database"
	"msu-forum/models"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Registro de usuário
func Register(c *fiber.Ctx) error {
	var data struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Phone    string `json:"phone"`
		Wallet   string `json:"wallet"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	// Verifica se cargo é válido
	validRoles := map[string]bool{
		"Admin": true, "Streamer": true, "Moderator": true, "Member": true,
	}
	if !validRoles[data.Role] {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role"})
	}

	// Criptografa senha
	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), 12)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error encoding password"})
	}

	// Insere no banco
	query := `INSERT INTO users (name, email, password, role, phone, wallet) 
			  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	var user models.User
	err = database.DB.QueryRow(
		query,
		data.Name, data.Email, string(hash), data.Role, data.Phone, data.Wallet,
	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error registering user"})
	}

	user.Name = data.Name
	user.Email = data.Email
	user.Role = data.Role
	user.Phone = data.Phone
	user.Wallet = data.Wallet

	return c.Status(201).JSON(user)
}

// Login de usuário
func Login(c *fiber.Ctx) error {
	var data struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	var user models.User
	err := database.DB.Get(&user, "SELECT * FROM users WHERE email=$1", data.Email)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credencials"})
	}

	// Verifica senha
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password)) != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credencials"})
	}

	// Gera token JWT
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error generating token"})
	}

	return c.JSON(fiber.Map{"token": t})
}
