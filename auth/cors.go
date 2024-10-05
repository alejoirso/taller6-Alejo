package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware configura los encabezados de CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lista de orígenes permitidos
		originsPermitidos := []string{
			"http://localhost:5173",
			"http://localhost:5174",
		}

		// Obtenemos el origen de la solicitud
		origin := c.Request.Header.Get("Origin")

		// Verificamos si el origen está en la lista de permitidos
		for _, origenPermitido := range originsPermitidos {
			if origin == origenPermitido {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// Configuración de CORS
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
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
