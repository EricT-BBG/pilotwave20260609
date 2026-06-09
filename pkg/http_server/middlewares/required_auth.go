package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func RequiredAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		accessToken := c.Request.Header.Get("Authentication")
		if len(accessToken) == 0 {
			accessToken = c.DefaultQuery("token", "")
		}

		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {

			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			secretKey := viper.GetString("auth.secret")

			return []byte(secretKey), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})

			c.Abort()
			return
		}

		// Get data in the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})

			c.Abort()
			return
		}

		// Initializing variables
		c.Set("uid", claims["uid"])
		c.Set("username", claims["username"])

		c.Next()
	}

}
