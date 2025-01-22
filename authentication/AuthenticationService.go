package authentication

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

var Secret = []byte("supersecretkeythatisverylongandrandom") // Replace <jwt-secret> with your secret key that is private to you.


func BuildJwtToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	tokenString, err := token.SignedString([]byte(Secret))
	if err != nil {
		log.Fatal("Error while generating token")
		return "", err
	}
	return tokenString, nil
}


func TokenMiddleware( next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Token is missing", http.StatusBadRequest)
			
			return
		}

		_, err := ParseToken(token[len("Bearer "):])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			
			return
		}

		next.ServeHTTP(w, r)
	})
}


func ParseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return Secret, nil
	})

	if err != nil {
		log.Fatal("Error while parsing token")
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Println(claims)
	} else {
		return nil, errors.New("invalid token")
	}

	return token, nil
}