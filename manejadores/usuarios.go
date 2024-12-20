package manejadores

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"taller6/auth"
	"taller6/base_datos"
	"taller6/modelos"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt" // Para encriptar contraseñas
)

// CrearUsuario maneja la creación de un nuevo usuario
func CrearUsuario(c *gin.Context) {
	var usuario modelos.UsuarioConToken

	// Validamos la entrada
	if err := c.ShouldBindJSON(&usuario); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos incorrectos"})
		return
	}

	// Verificar si el usuario o correo ya existen
	var existeUsuario int
	consultaVerificacion := `SELECT COUNT(*) FROM usuarios WHERE nombre_usuario = ? OR correo = ?`
	err := base_datos.BD.QueryRow(consultaVerificacion, usuario.NombreUsuario, usuario.Correo).Scan(&existeUsuario)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar la existencia del usuario"})
		return
	}
	if existeUsuario > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "El usuario ya existe."})
		return
	}

	// Encriptamos la contraseña antes de guardarla
	contrasenaEncriptada, err := bcrypt.GenerateFromPassword([]byte(usuario.Contrasena), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al encriptar la contraseña"})
		return
	}
	usuario.Contrasena = string(contrasenaEncriptada)
	usuario.CreadoEn = time.Now()

	// Insertamos el usuario en la base de datos
	consulta := `INSERT INTO usuarios (nombre_usuario, correo, contrasena, creado_en) VALUES (?, ?, ?, ?)`
	resultado, err := base_datos.BD.Exec(consulta, usuario.NombreUsuario, usuario.Correo, usuario.Contrasena, usuario.CreadoEn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el usuario, no logra guardarse en la base."})
		return
	}

	// Obtener el ID del usuario recién insertado
	usuarioID, err := resultado.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo obtener el ID del usuario"})
		return
	}
	usuario.ID = uint(usuarioID)

	// Generar el token para el usuario
	token, err := auth.GenerarToken(usuario.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo generar el token"})
		return
	}

	usuario.Token = token

	// Crear la respuesta sin la contraseña
	usuarioSinContrasena := modelos.UsuarioConToken{
		ID:            usuario.ID,
		NombreUsuario: usuario.NombreUsuario,
		Correo:        usuario.Correo,
		CreadoEn:      usuario.CreadoEn,
		Token:         token,
	}

	// Devolvemos el usuario creado (sin la contraseña)
	c.JSON(http.StatusCreated, usuarioSinContrasena)
}

// Login maneja la autenticación de un usuario
func Login(c *gin.Context) {
	var datosLogin struct {
		NombreUsuario string `json:"nombre_usuario"`
		Contrasena    string `json:"contrasena"`
	}

	// Validamos la entrada, es decir, valido que me envien los campos usuario y contraseña. Si me mandan otro campo, ROMPE (400 Bad Request)
	if err := c.ShouldBindJSON(&datosLogin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos incorrectos"})
		return
	}

	//Le digo traeme id, nombre, contraseña de la tabla usuario donde el nombre de usuario sea el nombre_usuario que me envia el usuario
	var usuario modelos.Usuario
	consulta := `SELECT id, nombre_usuario, contrasena FROM usuarios WHERE nombre_usuario = ?`
	//queryRow ejecuta la consulta y devuelve una fila \ Scan asigna los valores a la fila que devolvio queryRow (en este caso, id, nombre, contraseña del usuario)
	err := base_datos.BD.QueryRow(consulta, datosLogin.NombreUsuario).Scan(&usuario.ID, &usuario.NombreUsuario, &usuario.Contrasena)
	//sino encuentra ese nombre de usuario en la base, ROMPE (401 Unauthorized)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "El usuario ingresado no existe."})
		return
	}

	// Log propio para ver si va todo OK hasta aca
	log.Printf("(Login) Contraseña proporcionada en Login es: %s", datosLogin.Contrasena)
	log.Printf("(Login) Contraseña encriptada en DB: %s", usuario.Contrasena)

	// **************Comparamos la contraseña encriptada*****************************
	// usuario.Contrasena es la contraseña encriptada en la base
	// datosLogin.Contrasena es la contraseña en texto plano que envia el usuario
	//CompareHashAndPassword realiza la comparacion entre la contraseña hasheada guardada en la base con la contraseña que envia el usuario luego de hashearla
	//Si estas dos no coinciden, ROMPE (401 Unauthorized)
	if err := bcrypt.CompareHashAndPassword([]byte(usuario.Contrasena), []byte(datosLogin.Contrasena)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "La contraseña hasheada guardada en la base no coincide con la contraseña que envió el usuario, hasheada en el momento."})
		return
	}

	// Generamos el token JWT
	//x := strconv.FormatUint(uint64(usuario.ID), 10)
	token, err := auth.GenerarToken(usuario.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo generar el token"})
		return
	}

	// Enviamos el token al cliente
	c.JSON(http.StatusCreated, gin.H{"token": token}) //201

}

// ObtenerUsuario permite obtener el perfil del usuario o de otro si es administrador
func ObtenerUsuario(c *gin.Context) {
	esAdmin, existe := c.Get("es_admin")
	if !existe {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo verificar el tipo de usuario"})
		return
	}

	if esAdmin.(bool) {
		id := c.Param("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido (es admin)"})
			return
		}
		ObtenerUsuarioPorID(c, idInt)
	} else {
		id := c.GetString("id_usuario")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido (no es admin)"})
			return
		}
		ObtenerUsuarioPorID(c, idInt)
	}
}

// obtener me
func ObtenerUsuarioPorID(c *gin.Context, id int) {
	var usuario modelos.Usuario
	var creadoEn string // Usamos string para capturar el valor de la fecha

	consulta := `SELECT id, nombre_usuario, correo, creado_en FROM usuarios WHERE id = ?`
	err := base_datos.BD.QueryRow(consulta, id).Scan(&usuario.ID, &usuario.NombreUsuario, &usuario.Correo, &creadoEn)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("ObtenerUsuario: No se encontraron filas para el ID:", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "ObtenerUsuario: Usuario no encontrado"})
		} else {
			log.Println("Error al ejecutar la consulta:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar el usuario"})
		}
		return
	}

	// Convertimos la cadena a time.Time porque en uint rompia
	usuario.CreadoEn, err = time.Parse("2006-01-02 15:04:05", creadoEn)
	if err != nil {
		log.Println("Error al convertir la fecha:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar la fecha de creación"})
		return
	}

	c.JSON(http.StatusOK, usuario)
}

// ActualizarUsuario maneja la actualización de un usuario
func ActualizarUsuario(c *gin.Context) {
	esAdmin, existe := c.Get("es_admin")
	if !existe {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo verificar el tipo de usuario"})
		return
	}

	// Obtener datos enviados por el cliente
	var datosUsuario modelos.Usuario
	if err := c.ShouldBindJSON(&datosUsuario); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos incorrectos"})
		return
	}

	// Determinar ID según si es admin o no
	var id string
	if esAdmin.(bool) {
		id = c.Param("id")
	} else {
		id = c.GetString("id_usuario")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Construir consulta de actualización solo con campos presentes
	consulta := "UPDATE usuarios SET "
	args := []interface{}{}

	if datosUsuario.NombreUsuario != "" {
		consulta += "nombre_usuario = ?, "
		args = append(args, datosUsuario.NombreUsuario)
	}
	if datosUsuario.Correo != "" {
		consulta += "correo = ?, "
		args = append(args, datosUsuario.Correo)
	}
	if datosUsuario.Contrasena != "" {
		contrasenaEncriptada, err := bcrypt.GenerateFromPassword([]byte(datosUsuario.Contrasena), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al encriptar la contraseña"})
			return
		}
		consulta += "contrasena = ?, "
		args = append(args, string(contrasenaEncriptada))
	}

	// Eliminar la última coma y espacio
	consulta = consulta[:len(consulta)-2]
	consulta += " WHERE id = ?"
	args = append(args, idInt)

	// Log para verificar la consulta antes de ejecutarla
	log.Printf("Consulta de actualización: %s, con parámetros: %v", consulta, args)

	// Ejecutar la consulta
	_, err = base_datos.BD.Exec(consulta, args...)
	if err != nil {
		// Log del error para mayor detalle
		log.Printf("Error al ejecutar la consulta de actualización: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar el usuario"})
		return
	}

	// Recuperar los datos actualizados del usuario, excluyendo la contraseña
	var usuarioActualizado modelos.UsuarioSinContrasena
	consulta = "SELECT id, nombre_usuario, correo, creado_en FROM usuarios WHERE id = ?"
	row := base_datos.BD.QueryRow(consulta, idInt)

	// Utilizar sql.NullString para manejar la fecha como string
	var creadoEn []byte
	if err := row.Scan(&usuarioActualizado.ID, &usuarioActualizado.NombreUsuario, &usuarioActualizado.Correo, &creadoEn); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error al recuperar los datos del usuario: %v", err)})
		}
		return
	}

	// Convertir la fecha de []byte a time.Time
	parsedDate, err := time.Parse("2006-01-02 15:04:05", string(creadoEn))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al parsear la fecha de creación"})
		return
	}

	usuarioActualizado.CreadoEn = parsedDate

	// Devolver el usuario actualizado sin la contraseña
	c.JSON(http.StatusOK, usuarioActualizado)
}

/* ******************************************************
// ActualizarUsuario maneja la actualización de un usuario
func ActualizarUsuario(c *gin.Context) {
	esAdmin, existe := c.Get("es_admin")
	if !existe {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo verificar el tipo de usuario"})
		return
	}

	// Obtener datos enviados por el cliente
	var datosUsuario modelos.Usuario
	if err := c.ShouldBindJSON(&datosUsuario); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos incorrectos"})
		return
	}

	// Determinar ID según si es admin o no
	var id string
	if esAdmin.(bool) {
		id = c.Param("id")
	} else {
		id = c.GetString("id_usuario")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Construir consulta de actualización solo con campos presentes
	consulta := "UPDATE usuarios SET "
	args := []interface{}{}

	if datosUsuario.NombreUsuario != "" {
		consulta += "nombre_usuario = ?, "
		args = append(args, datosUsuario.NombreUsuario)
	}
	if datosUsuario.Correo != "" {
		consulta += "correo = ?, "
		args = append(args, datosUsuario.Correo)
	}
	if datosUsuario.Contrasena != "" {
		contrasenaEncriptada, err := bcrypt.GenerateFromPassword([]byte(datosUsuario.Contrasena), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al encriptar la contraseña"})
			return
		}
		consulta += "contrasena = ?, "
		args = append(args, string(contrasenaEncriptada))
	}

	// Eliminar la última coma y espacio
	consulta = consulta[:len(consulta)-2]
	consulta += " WHERE id = ?"
	args = append(args, idInt)

	_, err = base_datos.BD.Exec(consulta, args...)
	if err != nil {
		log.Println("Error al actualizar el usuario:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar el usuario"})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"mensaje": "Usuario actualizado correctamente"})
}
*/

// EliminarUsuario borra un usuario por su ID
func EliminarUsuario(c *gin.Context) {
	id := c.Param("id")

	// Convertimos el ID de string a entero
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	consulta := `DELETE FROM usuarios WHERE id = ?`
	_, err = base_datos.BD.Exec(consulta, idInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo eliminar el usuario"})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"mensaje": "Usuario eliminado correctamente"})
}

// ObtenerUsuarios trae todos los usuarios de la base de datos sin validación de token
func ObtenerUsuarios(c *gin.Context) {
	var usuarios []modelos.Usuario
	rows, err := base_datos.BD.Query("SELECT id, nombre_usuario, correo, creado_en FROM usuarios")
	if err != nil {
		log.Println("Error al ejecutar la consulta:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar los usuarios"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var usuario modelos.Usuario
		var creadoEn string
		err := rows.Scan(&usuario.ID, &usuario.NombreUsuario, &usuario.Correo, &creadoEn)
		if err != nil {
			log.Println("Error al escanear fila:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar los datos de usuarios"})
			return
		}
		usuario.CreadoEn, err = time.Parse("2006-01-02 15:04:05", creadoEn)
		if err != nil {
			log.Println("Error al convertir la fecha:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar la fecha de creación"})
			return
		}
		usuarios = append(usuarios, usuario)
	}

	c.JSON(http.StatusOK, usuarios)
}
