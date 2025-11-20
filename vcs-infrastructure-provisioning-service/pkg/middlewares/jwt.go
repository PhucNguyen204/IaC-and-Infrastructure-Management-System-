package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTMiddleware struct {
	jwtSecret []byte
}

func NewJWTMiddleware(jwtSecret string) *JWTMiddleware {
	return &JWTMiddleware{
		jwtSecret: []byte(jwtSecret),
	}
}

func (m *JWTMiddleware) CheckBearerAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Printf("JWT: Invalid signing method: %v\n", token.Method)
			return nil, jwt.ErrSignatureInvalid
		}
		secretLen := len(m.jwtSecret)
		if secretLen > 10 {
			secretLen = 10
		}
		fmt.Printf("JWT: Using secret (first 10 chars): %s\n", string(m.jwtSecret[:secretLen]))
		return m.jwtSecret, nil
	})

	if err != nil {
		fmt.Printf("JWT Parse Error: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
		return
	}
	
	if !token.Valid {
		fmt.Printf("JWT: Token is not valid\n")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
		return
	}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, ok := claims["sub"].(string); ok {
				c.Set("user_id", userID)
			}
		}

		c.Next()
	}
}

