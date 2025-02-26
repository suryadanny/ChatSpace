package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// ValidateLoginRequestMiddleWare is a middleware function that validates the signin request
func ValidateUserRequestMiddleWare(chain http.HandlerFunc) http.HandlerFunc {
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		request := make(map[string]string)

		request_body , err := io.ReadAll(r.Body)

		if err != nil {
			log.Println("error while reading request body :",err)
			w.WriteHeader(http.StatusBadRequest)
		}
		
		un_err := json.Unmarshal(request_body, &request)
	
		if un_err != nil {
			log.Println("error while unmarshalling request body: ",un_err )
			w.WriteHeader(http.StatusBadRequest)
		}
	

		if request["password"] == "" || request["email"] == ""  || request["user_name"] == "" {
			http.Error(w, "email or password is missing", http.StatusBadRequest)
			
			return 
		}

		r.Body = io.NopCloser(bytes.NewReader(request_body))

		chain.ServeHTTP(w, r)


	})


	
}

//checks if the request has the required fields for login
func ValidateLoginRequestMiddleWare(chain http.HandlerFunc) http.HandlerFunc {
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := make(map[string]string)

		request_body , err := io.ReadAll(r.Body)

		if err != nil {
			log.Println("error while reading request body")
			w.WriteHeader(http.StatusBadRequest)
		}
		
		un_err := json.Unmarshal(request_body, &request)
	
		if un_err != nil {
			log.Println("error while unmarshalling request body")
			w.WriteHeader(http.StatusBadRequest)
		}
	
		username := request["user_name"]
		password := request["password"]
	if username == "" ||  password == "" {
		http.Error(w, "username or password is missing", http.StatusBadRequest)
		return 
	}

	r.Body = io.NopCloser(bytes.NewReader(request_body))

	chain.ServeHTTP(w, r)


	})	
}


// func ValidateTokenMiddleWare(chain http.HandlerFunc) http.HandlerFunc {

// }