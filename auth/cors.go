package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware configura los encabezados de CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Permitir el origen de la solicitud
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// Configuración de CORS
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Manejo del preflight (opciones)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// Continuamos con la solicitud
		c.Next()
	}
}
