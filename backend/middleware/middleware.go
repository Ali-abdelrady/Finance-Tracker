package middleware

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// المفتاح السري لتوقيع التوكن (يُفضل حفظه في Environment Variables)
var jwtSecret = []byte(func() string {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}
	return "your_super_secret_key_change_in_production"
}())

// Custom Claims لتخزين البيانات داخل التوكن
type JWTPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// 1. دالة لتوليد التوكن عند تسجيل الدخول بنجاح
func GenerateToken(userID, email string) (string, error) {
	claims := JWTPayload{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // الصلاحية لـ 24 ساعة
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 2. الـ Middleware لحماية الـ Endpoints
func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or malformed JWT token",
			})
		}

		// التأكد من أن الصيغة تبدأ بـ Bearer
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization header format. Use 'Bearer <token>'",
			})
		}

		tokenStr := parts[1]

		// التحقق من صحة التوكن
		claims := &JWTPayload{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			// التأكد من أن خوارزمية التوقيع هي نفسها (HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// تخزين الـ user_id في الـ Context للوصول إليه مباشرة داخل الـ Handlers
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}
