package auth

import (
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var claveJWT = []byte("irso2024") // Clave secreta para firmar los tokens
const adminID = 1                 // Asumimos que el ID del admin es 1

// Reclamos define lo que contendrá el token
type Reclamos struct {
	Id uint `json:"Id"`
	jwt.RegisteredClaims
}

// GenerarToken crea un token para el usuario
func GenerarToken(id_usuario uint) (string, error) {
	// Definimos los reclamos del token (información que contendrá)
	reclamos := Reclamos{
		Id: id_usuario,
	}

	// Si el usuario no es "admin", establecer tiempo de expiración
	if id_usuario != adminID {
		reclamos.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)) // El token expira en 1 semana
	}

	// Creamos el token con el método de firma HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, reclamos)

	// Firmamos el token con la clave secreta
	tokenFirmado, err := token.SignedString(claveJWT)
	if err != nil {
		return "", err
	}

	return tokenFirmado, nil
}

// ValidarToken verifica que el token es válido
func ValidarToken(tokenString string) (*Reclamos, error) {
	// Parseamos el token y lo validamos
	token, err := jwt.ParseWithClaims(tokenString, &Reclamos{}, func(token *jwt.Token) (interface{}, error) {
		return claveJWT, nil
	})

	if err != nil {
		return nil, err
	}

	// Verificamos que el token sea válido y que los reclamos estén presentes
	if reclamos, ok := token.Claims.(*Reclamos); ok && token.Valid {
		return reclamos, nil
	}

	// Si llegamos aquí, el token no es válido
	log.Println("ValidarToken: token inválido o reclamos incorrectos")
	return nil, errors.New("ValidarToken: token inválido o reclamos incorrectos")
}
