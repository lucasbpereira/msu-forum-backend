package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"msu-forum/database"
	"msu-forum/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// --- Constantes para melhorar a legibilidade e manutenção ---
const (
	msuAPIKeyHeader  = "x-nxopen-api-key"
	msuAPIBaseURL    = "https://openapi.msu.io/v1beta/accounts/%s/characters?paginationParam.pageNo=1"
	defaultUserRole  = "Member"
	jwtSecretEnvKey  = "JWT_SECRET"
	msuAPIKeyEnvKey  = "MSU_API_KEY"
	tokenDuration    = 24 * time.Hour
	apiClientTimeout = 10 * time.Second
)

// --- Estruturas para a API Externa (sem alteração) ---
type APICharacterData struct {
	Level    int    `json:"level"`
	ImageURL string `json:"imageUrl"`
}

type APICharacter struct {
	Name string           `json:"name"`
	Data APICharacterData `json:"data"`
}

type APIResponse struct {
	Characters []APICharacter `json:"characters"`
}

// =============================================================================
// HANDLERS (Controladores de Rota)
// =============================================================================

// Register lida com o registro de um novo usuário.
func Register(c *fiber.Ctx) error {
	var req struct {
		Wallet string `json:"wallet"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}
	if req.Wallet == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Wallet é obrigatória"})
	}

	// 1. Verificar se o usuário já existe no banco de dados local.
	_, err := findUserByWallet(req.Wallet)
	if err != sql.ErrNoRows { // Se o erro NÃO for "não encontrado", algo deu errado.
		if err == nil {
			return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "Wallet já cadastrada"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao verificar a wallet no banco de dados"})
	}

	// 2. Buscar dados do personagem na API externa.
	apiResponse, err := getMSUCharacterData(req.Wallet)
	if err != nil {
		// O erro de getMSUCharacterData já vem com o status code apropriado.
		return c.Status(fiber.ErrBadGateway.Code).JSON(fiber.Map{"error": err.Error()})
	}
	if len(apiResponse.Characters) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Nenhum personagem encontrado para esta wallet"})
	}

	// 3. Criar o novo usuário no banco de dados.
	firstCharacter := apiResponse.Characters[0]
	newUser, err := createNewUser(req.Wallet, firstCharacter.Name, firstCharacter.Data.ImageURL)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao registrar usuário", "details": err.Error()})
	}

	// 4. Gerar o token JWT.
	tokenString, err := generateJWT(newUser)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao gerar token"})
	}

	// 5. Retornar o usuário criado e o token.
	cookie := new(fiber.Cookie)
	cookie.Name = "auth_token" // Nome do cookie
	cookie.Value = tokenString
	cookie.Expires = time.Now().Add(tokenDuration)
	cookie.HTTPOnly = true // Essencial: impede o acesso via JavaScript
	// cookie.Secure = true // Para produção: envie apenas via HTTPS. Comente em dev local com HTTP.
	cookie.SameSite = "Strict" // Ou "Lax"

	// Define o cookie na resposta
	c.Cookie(cookie)

	// 5. Retornar o usuário criado, mas SEM o token no corpo.
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		// O campo "token" foi removido daqui
		"user": newUser, // É uma boa prática retornar o usuário criado
	})
}

// Login lida com a autenticação de um usuário existente.
func Login(c *fiber.Ctx) error {
	var req struct {
		Wallet string `json:"wallet"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}
	if req.Wallet == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Wallet é obrigatória"})
	}

	// 1. Encontrar usuário no banco de dados.
	user, err := findUserByWallet(req.Wallet)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Wallet não encontrada"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao buscar usuário"})
	}

	if !user.IsActive {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Usuário inativo"})
	}

	// 2. Atualizar o 'last_seen' do usuário (pode ser em goroutine se virar gargalo).
	updateLastSeen(user.ID)

	// 3. (Opcional, mas mantido da sua lógica) Buscar dados frescos da API externa.
	// Isso pode ser útil para atualizar o avatar ou nome do usuário se ele mudar no jogo.
	apiResponse, err := getMSUCharacterData(req.Wallet)
	if err != nil {
		// Não falhamos o login se a API externa estiver fora, mas podemos logar o erro.
		fmt.Printf("Aviso: Falha ao buscar dados externos para a wallet %s: %v\n", req.Wallet, err)
	}

	// 4. Gerar o token JWT.
	tokenString, err := generateJWT(user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao gerar token"})
	}

	// 5. Retornar os dados do usuário e o token.
	cookie := new(fiber.Cookie)
	cookie.Name = "auth_token" // Nome do cookie
	cookie.Value = tokenString
	cookie.Expires = time.Now().Add(tokenDuration)
	cookie.HTTPOnly = true // Essencial: impede o acesso via JavaScript
	// cookie.Secure = true // Para produção: envie apenas via HTTPS. Comente em dev local com HTTP.
	cookie.SameSite = "Strict" // Ou "Lax"

	// Define o cookie na resposta
	c.Cookie(cookie)

	// 5. Retornar o usuário criado, mas SEM o token no corpo.
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"user":       user,
		"characters": apiResponse.Characters,
	})
}

func Logout(c *fiber.Ctx) error {
	// Cria um cookie com o mesmo nome, mas com data de expiração no passado.
	// Isso instrui o navegador a removê-lo.
	cookie := new(fiber.Cookie)
	cookie.Name = "auth_token"
	cookie.Value = ""                           // O valor não importa
	cookie.Expires = time.Now().Add(-time.Hour) // Data no passado
	cookie.HTTPOnly = true
	// cookie.Secure = true // Mantenha os mesmos atributos do cookie de login
	cookie.SameSite = "Strict"

	c.Cookie(cookie)

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Logout realizado com sucesso",
	})
}

// =============================================================================
// FUNÇÕES AUXILIARES (Lógica de Negócio e Acesso a Dados)
// =============================================================================

// getMSUCharacterData encapsula a chamada à API externa da MSU.
func getMSUCharacterData(wallet string) (*APIResponse, error) {
	apiKey := os.Getenv(msuAPIKeyEnvKey)
	if apiKey == "" {
		return nil, fmt.Errorf("a chave da API MSU não está configurada")
	}

	apiURL := fmt.Sprintf(msuAPIBaseURL, wallet)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro interno ao criar requisição: %w", err)
	}
	req.Header.Add(msuAPIKeyHeader, apiKey)

	client := &http.Client{Timeout: apiClientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha na comunicação com o serviço externo: %w", err)
	}
	defer resp.Body.Close()

	// Log para depuração
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("API Externa [Wallet: %s, Status: %d], Corpo: %s\n", wallet, resp.StatusCode, string(body))

	// Recriar o corpo para que possa ser lido pelo decoder
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wallet não encontrada ou inválida no serviço externo (status: %d)", resp.StatusCode)
	}

	var apiResponse APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar a resposta do serviço externo: %w", err)
	}

	return &apiResponse, nil
}

// findUserByWallet busca um usuário no banco de dados pela sua wallet.
func findUserByWallet(wallet string) (*models.User, error) {
	var user models.User
	query := "SELECT * FROM users WHERE wallet = $1"
	err := database.DB.Get(&user, query, wallet)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// createNewUser insere um novo usuário no banco de dados.
func createNewUser(wallet, username, avatarURL string) (*models.User, error) {
	var user models.User
	now := time.Now()
	query := `
        INSERT INTO users (wallet, username, role, reputation, is_active, created_at, last_seen, avatar_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, wallet, username, role, reputation, is_active, created_at, last_seen, avatar_url
    `
	err := database.DB.QueryRow(
		query, wallet, username, defaultUserRole, 0, true, now, now, avatarURL,
	).Scan(
		&user.ID, &user.Wallet, &user.Username, &user.Role, &user.Reputation,
		&user.IsActive, &user.CreatedAt, &user.LastSeen, &user.AvatarURL,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// generateJWT cria um novo token JWT para o usuário.
func generateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"role":     user.Role,
		"wallet":   user.Wallet,
		"username": user.Username,
		"exp":      time.Now().Add(tokenDuration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv(jwtSecretEnvKey)))
}

// updateLastSeen atualiza o campo last_seen para um usuário.
func updateLastSeen(userID int) {
	query := "UPDATE users SET last_seen=$1 WHERE id=$2"
	_, err := database.DB.Exec(query, time.Now(), userID)
	if err != nil {
		// Em um sistema real, aqui você logaria o erro, mas não bloquearia o login.
		fmt.Printf("Erro ao atualizar last_seen para o usuário ID %d: %v\n", userID, err)
	}
}
