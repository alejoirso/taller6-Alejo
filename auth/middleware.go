package auth

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequiereAutenticacion es un middleware que verifica que el usuario tenga un token válido
func RequiereAutenticacion() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Configuración de CORS
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173") // Cambia el dominio si es necesario
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Manejo del preflight (opciones)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// Obtenemos el token del header Authorization
		tokenString := c.GetHeader("Authorization")

		// Verificamos si el token fue proporcionado
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token no proporcionado"})
			c.Abort()
			return
		}

		// Comprobamos si el formato es "Bearer <token>"
		if !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de token inválido"})
			c.Abort()
			return
		}

		// Extraemos el token sin la palabra "Bearer "
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Imprimir el token que estamos recibiendo
		log.Println("Token recibido:", tokenString)

		// Validamos el token
		usuario, err := ValidarToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		// Imprimir detalles del usuario
		log.Println("Usuario ID:", usuario.Id)

		// Guardamos el id del usuario en el contexto para futuras solicitudes
		c.Set("id_usuario", strconv.Itoa(int(usuario.Id))) // Convertir uint a int y luego a string

		// Verificamos si el usuario es administrador (ID 1)
		if usuario.Id == 1 {
			log.Println("Acceso otorgado al usuario administrador")
			c.Set("es_admin", true) // Guardamos una marca en el contexto de que SI es admin
		} else {
			c.Set("es_admin", false)
		}

		c.Next() // Continuamos la ejecución si el token es válido
	}
}
