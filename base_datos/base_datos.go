// 3
package base_datos

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

const url = "root:De$boca10@tcp(localhost:3306)/taller6"

//ruta de mi base, usuario y contraseña, + lo que sigue (localhost:3306 - es donde esta el servidor sql -) y luego el nombre de la base

// guarda la conexion a la base
var BD *sql.DB //variable global de tipo sql.DB

// funcion que realiza la conexion a la base
func ConectarBD() {
	conexion, err := sql.Open("mysql", url)
	//nombre del driver + la ruta de la base (guardado en la constante "url" previamente)
	//esto devuele la conexion + un error. Por eso procedo a capturarlos y verificar con un if lo que ocurrio
	if err != nil {
		panic(err)
	}
	fmt.Println("Conexion a la base de datos exitosa")
	BD = conexion //guardo la cone en la variable global "db"
}

// funcion para cerrar la conexion a la base
func CerrarBD() {
	BD.Close() //gracias a la cone guardada previamente, procedo a cerrarla
}

// Funcion para verificar si la conexion con la base de datos continua
func Ping() {
	if err := BD.Ping(); err != nil { //si sigue OK no hace nada, si pincho devuelve el error
		panic(err)
	}
}

// Verificar si existe una tabla
func TablaExistente(nombreTabla string) bool {
	//Le paso el nombre de la tabla, y si existe devuelve "true" sino devuelve "false"
	sql := fmt.Sprintf("SHOW TABLES LIKE '%s'", nombreTabla)

	rows, err := BD.Query(sql) //Query: recibe una consulta SQL y argumentos indefinidos
	//db.Query retorna un rows y un error, entonces los capturo.
	if err != nil {
		fmt.Println("Error: ", err)
	}
	return rows.Next() //esto recorre la tabla, si puede recorrerla significa que existe, devuelve "true" y sino "false"

}

// Crea una tabla segun el esquema que le indiquemos como parametro.
func CrearTabla(schema string, nombre string) {
	//Sino existe la tabla, creala
	if !TablaExistente(nombre) {
		_, err := BD.Exec(schema) //ejecuta un sql
		//maneja un posible error al crear una tabla
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

}

// Crear usuario administrador si no existe
func CrearUsuarioAdmin() {
	var id int
	err := BD.QueryRow("SELECT id FROM usuarios WHERE nombre_usuario = 'admin'").Scan(&id)
	// Si el usuario 'admin' no existe, lo creamos
	if err == sql.ErrNoRows {
		// Hasheamos la contraseña del administrador
		contrasenaAdmin := "admin123"
		contrasenaEncriptada, err := bcrypt.GenerateFromPassword([]byte(contrasenaAdmin), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Error al encriptar la contraseña de admin: %v", err)
		}

		// Creamos el usuario administrador
		consulta := `INSERT INTO usuarios (nombre_usuario, correo, contrasena, creado_en) VALUES ('admin', '', ?, ?)`
		_, err = BD.Exec(consulta, string(contrasenaEncriptada), time.Now())
		if err != nil {
			log.Fatalf("Error al crear el usuario administrador: %v", err)
		} else {
			fmt.Println("Usuario administrador creado exitosamente")
		}
	}
}
