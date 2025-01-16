package service

import (
	"dev/chatspace/dbservice"
	"dev/chatspace/models"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)


type UserService struct {
}


func NewUserService() *UserService{
	return &UserService{}
}

func (u *UserService) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := dbservice.GetAllUsers()

	if err != nil {
		log.Fatal("error occurred while fetching users")
	}

	serUser , _ := json.Marshal(users)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(serUser))
}


func (u *UserService) GetUser(w http.ResponseWriter, r *http.Request) {
	id  := chi.URLParam(r, "id")

	user, err := dbservice.GetUser(id)

	if err != nil {
		log.Fatal("error occurred while fetching users")
	}

	serUser , _ := json.Marshal(user)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(serUser))
}

func (u *UserService)CreateUser(w http.ResponseWriter, r *http.Request) {

	user := &models.User{}
	request_body , err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("error while reading request body")
	    w.WriteHeader(http.StatusBadRequest)
		return
	}

	var err_un = json.Unmarshal(request_body, user)

	if err_un != nil {
		log.Println("error while unmarshalling request body")
	    w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = dbservice.InsertNewUser(user)

	if err != nil {	
		log.Println("error while creating user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	

}
