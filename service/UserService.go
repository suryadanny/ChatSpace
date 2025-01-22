package service

import (
	"dev/chatspace/authentication"
	"dev/chatspace/dbservice"
	"dev/chatspace/models"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"time"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func (u *UserService) Login(w http.ResponseWriter, r *http.Request) {

	
	request := make(map[string]string)

	request_body , err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("error in retrieving request body")
		w.WriteHeader(http.StatusBadRequest)
	}
	
	un_err := json.Unmarshal(request_body, &request)

	if un_err != nil {
		log.Println("error while unmarshalling request body ", un_err)
		w.WriteHeader(http.StatusBadRequest)
	}

	username := request["user_name"]
	password := request["password"]
	log.Println("username : ", username)

	user, err := u.userRepo.GetUserByField("user_name", username)

	if err != nil {
		log.Println("error while fetching user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.Password != password {
		log.Println("password is incorrect")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"username": user.UserName,
		"email": user.Email,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	token_string, err := authentication.BuildJwtToken(claims)

	if err != nil {
		log.Println("error while generating token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer " + token_string)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("login successful"))

}
