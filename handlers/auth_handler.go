package handlers

import (
	"fmt"
	"os"
	"time"

	"msu-forum/database"
	"msu-forum/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Registro de usuário
// Registro de usuário
func Register(c *fiber.Ctx) error {
	var data struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
		Wallet   string `json:"wallet"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	// Validação de campos obrigatórios (Wallet agora é obrigatória)
	if data.Username == "" || data.Email == "" || data.Password == "" || data.Wallet == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Username, email, password e wallet são obrigatórios"})
	}

	// Verifica se a wallet já existe
	var existingUser models.User
	walletCheckQuery := "SELECT id FROM users WHERE wallet = $1"
	err := database.DB.Get(&existingUser, walletCheckQuery, data.Wallet)
	if err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Wallet já está em uso"})
	}

	// Criptografa senha
	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao codificar senha"})
	}

	// Define avatar URL padrão
	avatarURL := "/assets/avatar_default.png"

	// Insere no banco
	var user models.User
	now := time.Now()
	query := `
		INSERT INTO users 
			(username, email, password, role, phone, wallet, reputation, is_active, created_at, last_seen, avatar_url)
		VALUES 
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, username, email, role, phone, wallet, reputation, is_active, created_at, last_seen, avatar_url
	`

	err = database.DB.QueryRow(
		query,
		data.Username,
		data.Email,
		string(hash),
		"Member",
		data.Phone,
		data.Wallet, // Agora obrigatório
		0,           // reputation
		true,        // is_active
		now,
		now,
		avatarURL, // Avatar URL padrão
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.Phone,
		&user.Wallet,
		&user.Reputation,
		&user.IsActive,
		&user.CreatedAt,
		&user.LastSeen,
		&user.AvatarURL,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao registrar usuário", "details": err.Error()})
	}

	return c.Status(201).JSON(user)
}

func Login(c *fiber.Ctx) error {
	var data struct {
		Wallet   string `json:"wallet"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	// Validação de campo obrigatório
	if data.Wallet == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Wallet é obrigatória"})
	}

	var user models.User
	query := "SELECT * FROM users WHERE wallet = $1"

	err := database.DB.Get(&user, query, data.Wallet)
	if err != nil {
		fmt.Printf("Erro na query: %v\n", err)
		fmt.Println("Nenhum usuário encontrado com essa wallet.")
		return c.Status(401).JSON(fiber.Map{"error": "Credenciais inválidas"})
	}

	fmt.Printf("Usuário encontrado: ID=%d, Wallet=%s\n", user.ID, user.Wallet)
	fmt.Printf("Hash no banco: %s\n", user.Password)

	// Teste a senha manualmente para debug
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
	if err != nil {
		fmt.Printf("Erro na comparação de senha: %v\n", err)
		fmt.Println("Senha não confere")
		return c.Status(401).JSON(fiber.Map{"error": "Credenciais inválidas"})
	}

	fmt.Println("Login bem-sucedido!")

	// Verifica se usuário está ativo
	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "Usuário inativo"})
	}

	// Atualiza last_seen
	_, _ = database.DB.Exec("UPDATE users SET last_seen=$1 WHERE id=$2", time.Now(), user.ID)

	// Gera token JWT
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"wallet":  user.Wallet, // Inclui a wallet no token se necessário
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao gerar token"})
	}

	return c.JSON(fiber.Map{"token": t})
}
