package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tokha04/todo-list-api/tokens"
)

func Authentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientToken := ctx.Request.Header.Get("token")
		if clientToken == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "no token provided"})
			ctx.Abort()
			return
		}

		claims, err := tokens.ValidateToken(clientToken)
		if err != "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "could not validate token"})
			ctx.Abort()
			return
		}

		ctx.Set("user_id", claims.ID)
		ctx.Next()
	}
}
