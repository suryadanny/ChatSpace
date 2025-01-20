package service

import (
	"dev/chatspace/dbservice"
	"dev/chatspace/models"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/go-chi/chi"
)


type UserService struct {
	userRepo *dbservice.UserRepository
}


func NewUserService(userRepo *dbservice.UserRepository) *UserService{
	return &UserService{userRepo:userRepo}
}

func (u *UserService) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := u.userRepo.GetAllUsers()

	if err != nil {
		log.Fatal("error occurred while fetching users")
	}

	serUser , _ := json.Marshal(users)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(serUser))
}


func (u *UserService) GetUser(w http.ResponseWriter, r *http.Request) {
	id  := chi.URLParam(r, "id")

	user, err := u.userRepo.GetUser(id)

	if err != nil {
		log.Println("error occurred while fetching user")
	}

	serUser , _ := json.Marshal(user)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(serUser))
}

func (u *UserService) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id  := chi.URLParam(r, "id")

	err := u.userRepo.DeleteUser(id)

	if err != nil {
		log.Println("error occurred while fetching user")
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (u *UserService) UpdateUser(w http.ResponseWriter, r *http.Request) {
	user := make(map[string]string)
	user_id := chi.URLParam(r, "id")
	request_body , err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("error while reading request body")
	    w.WriteHeader(http.StatusBadRequest)
		return
	}

	var err_un = json.Unmarshal(request_body, &user)

	if err_un != nil {
		log.Println("error while unmarshalling request body")
	    w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = u.userRepo.UpdateUser(user,user_id)

	if err != nil {
		log.Println("error while updating user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	user.UserId = uuid.New().String()

	err = u.userRepo.CreateUser(user)

	if err != nil {	
		log.Println("error while creating user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	

}
