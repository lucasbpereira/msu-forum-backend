package middleware

import (
	"os"
	// "strings" // Não é mais necessário

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(c *fiber.Ctx) error {
	// 1. Obter o token do cookie chamado "auth_token"
	tokenString := c.Cookies("auth_token")

	// 2. Verificar se o cookie existe
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token ausente"})
	}

	// 3. O resto da lógica de validação do token permanece a mesma
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validação do método de assinatura
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Método de assinatura inesperado")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token inválido"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Claims do token inválidas"})
	}

	// Conversão correta do user_id
	userID := int(claims["user_id"].(float64))
	role := claims["role"].(string)

	c.Locals("user_id", userID)
	c.Locals("role", role)

	return c.Next()
}
