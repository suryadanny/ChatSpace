package authentication

import (
	"dev/chatspace/utils"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt/v5"
)

var Secret = []byte("supersecretkeythatisverylongandrandom") // Replace <jwt-secret> with your secret key that is private to you.


func BuildJwtToken(claims jwt.MapClaims) (string, error) {
	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	tokenString, err := token.SignedString([]byte(Secret))
	if err != nil {
		log.Fatal("Error while generating token")
		return "", err
	}
	return tokenString, nil
}


func TokenMiddleware( next http.Handler) http.Handler {
	//returning a handler function, will be used jwt token authentication
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//checking if token is present in the header
		token := r.Header.Get(utils.Authorization)
		if token == "" {
			http.Error(w, "Token is missing", http.StatusBadRequest)
			
			return
		}
		
		user_id := chi.URLParam(r, "id")
		//exracts the claims from the token
		claims, err := ParseToken(token[len(utils.Bearer):])

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			
			return
		}else if claims["user_id"] != user_id {
			//returning unauthorized if the user_id in the token does not match the user_id in the url
			log.Println(claims)
			http.Error(w, "Unauthorized user", http.StatusUnauthorized)
			
			return

		}
		next.ServeHTTP(w, r)
	})
}


func ParseToken(tokenString string) (jwt.MapClaims, error) {

	//parses the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return Secret, nil
	})

	if err != nil {
		log.Println("Error while parsing token: ",err)
		return nil, err
	}
	//relevant claims are extracted from the token
	claims, ok := token.Claims.(jwt.MapClaims)

	if ok && token.Valid {
		log.Println(claims)
	} else {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}