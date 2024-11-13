package modelos

import "time"

// Estructura de cómo se va a componer un usuario.
type Usuario struct {
	ID            uint      `json:"id"`
	NombreUsuario string    `json:"nombre_usuario"`
	Correo        string    `json:"correo"`
	Contrasena    string    `json:"contrasena"` // No mostramos la contraseña
	CreadoEn      time.Time `json:"creado_en"`
}

type UsuarioConToken struct {
	ID            uint      `json:"id"`
	NombreUsuario string    `json:"nombre_usuario"`
	Correo        string    `json:"correo"`
	Contrasena    string    `json:"contrasena"` // No mostramos la contraseña
	CreadoEn      time.Time `json:"creado_en"`
	Token         string    `json:"token"`
}

type UsuarioSinContrasena struct {
	ID            uint      `json:"id"`
	NombreUsuario string    `json:"nombre_usuario"`
	Correo        string    `json:"correo"`
	CreadoEn      time.Time `json:"creado_en"`
}

// Esquema para crear la base de datos usuarios si es que no existe ya
const UsuariosSchema string = `CREATE TABLE usuarios (
    id SERIAL PRIMARY KEY,
    nombre_usuario VARCHAR(50) UNIQUE NOT NULL,
    correo VARCHAR(100) UNIQUE NOT NULL,
    contrasena TEXT NOT NULL,
    creado_en TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`
