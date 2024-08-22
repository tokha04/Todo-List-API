package helpers

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPasword string, givenPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPasword))
	if err != nil {
		return false
	}

	return true
}
