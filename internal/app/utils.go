package app

import (
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetToken(login string, tokenAuth *jwtauth.JWTAuth) string {
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_login": login})
	return tokenString
}
