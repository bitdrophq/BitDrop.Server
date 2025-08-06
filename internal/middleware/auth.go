// middleware/auth.go
package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Malformed token"})
			return
		}

		var token *jwt.Token
		var err error

		// Attempt verification with custom JWT secret
		customSecret := os.Getenv("JWT_SECRET")
		token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(customSecret), nil
		})

		// If custom JWT fails, try Supabase secret
		if err != nil || !token.Valid {
			supabaseSecret := os.Getenv("SUPABASE_JWT_SECRET")
			token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(supabaseSecret), nil
			})
		}

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Accept `user_id` (custom) or `sub` (Supabase)
		var userId string
		if val, ok := claims["user_id"].(string); ok && val != "" {
			userId = val
		} else if val, ok := claims["sub"].(string); ok && val != "" {
			userId = val
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing user ID in token"})
			return
		}

		// Check expiration
		if exp, ok := claims["exp"].(float64); ok && time.Now().Unix() > int64(exp) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			return
		}

		c.Set("userId", userId)
		c.Next()
	}
}
