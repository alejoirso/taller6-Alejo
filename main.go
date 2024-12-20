package main

import (
	"taller6/auth"
	"taller6/base_datos"
	"taller6/manejadores"
	"taller6/modelos"

	"github.com/gin-gonic/gin"
)

func main() {
	// Tratativas con la base de datos
	base_datos.ConectarBD()
	defer base_datos.CerrarBD() // Aseguramos que la base de datos se cierre solo cuando el programa termine

	// Creamos la tabla "usuarios" si no existe
	base_datos.CrearTabla(modelos.UsuariosSchema, "usuarios")
	// Creamos el usuario "admin" si no existe
	base_datos.CrearUsuarioAdmin()

	// Creamos la instancia del servidor de Gin
	servidor := gin.Default()

	// Aplicar el middleware de CORS a todas las rutas
	servidor.Use(auth.CORSMiddleware())

	// Definimos las rutas para el CRUD de usuarios
	servidor.POST("/usuarios", manejadores.CrearUsuario) // Ruta pública para crear usuario (sin autenticación)
	servidor.POST("/login", manejadores.Login)           // Ruta pública para login (sin autenticación)
	servidor.GET("/usuarios", manejadores.ObtenerUsuarios)

	// Grupo de rutas protegidas por el middleware de autenticación
	rutasProtegidas := servidor.Group("/")
	rutasProtegidas.Use(auth.RequiereAutenticacion()) // Aplica el middleware solo a estas rutas
	{
		// Ruta para obtener y actualizar el propio perfil
		rutasProtegidas.GET("/me", manejadores.ObtenerUsuario)
		rutasProtegidas.PATCH("/me", manejadores.ActualizarUsuario)
		// Rutas solo accesibles por admin
		rutasProtegidas.GET("/usuarios/:id", manejadores.ObtenerUsuario)
		rutasProtegidas.PATCH("/usuarios/:id", manejadores.ActualizarUsuario)
		rutasProtegidas.DELETE("/usuarios/:id", manejadores.EliminarUsuario)
		//rutasProtegidas.GET("/usuarios", manejadores.ObtenerUsuarios)

	}

	// Arrancamos el servidor en el puerto 8080
	servidor.Run("0.0.0.0:8080")
}
