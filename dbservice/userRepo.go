package dbservice

import (
	"dev/chatspace/models"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
)

type UserRepository struct {
	session *gocqlx.Session
	//lock *sync.RWMutex
}


func NewUserRepository(session *gocqlx.Session) *UserRepository {
	return &UserRepository{session}
}

func (u *UserRepository) CreateUser(user *models.User) error {
	
	err := u.session.Query(models.UserTable.Insert()).BindStruct(user).ExecRelease()
	
	if err != nil {
		log.Println("error while inserting user : ", err)
		return err
	}

	return nil	
}


func (u *UserRepository) GetAllUsers() ([]*models.User, error) {
	usersList := []*models.User{}
	query := u.session.Query("select * from user", nil)
	if err := query.Select(&usersList); err != nil {
		log.Println("error while fetching users : ", err)
		return nil, err
	}
	return usersList, nil
}


func (u *UserRepository) DeleteUser(userId string) error {
	err := u.session.Query(models.UserTable.Delete()).BindMap(qb.M{"user_id": userId}).ExecRelease()
	if err != nil {
		log.Println("error while deleting user : ", err)
		return err
	}
	return nil
}

func (u *UserRepository) UpdateUser(user map[string]string, user_id string) error {
	models.UserTable.SelectQuery(*u.session)
	// models.UserTable.UpdateQuery(*u.session).
	qb_map := qb.M{}
	update := qb.Update("store.user")
	for key, value := range user{
		update.Set(key)
		qb_map[key] = value
	}

	qb_map["user_id"] = user_id
	
	stmt, names := update.Where(qb.Eq("user_id")).ToCql()
	
	if err := u.session.Query(stmt, names).BindMap(qb_map).ExecRelease(); err != nil {
		log.Println("error while updating user : ", err)
		return err
	}

	return nil
}

func (u *UserRepository) GetUser(userId string) (*models.User, error) {
	user := &models.User{}
	err := u.session.Query(models.UserTable.Select()).BindMap(qb.M{"user_id": userId}).GetRelease(user)
	if err != nil {
		log.Printf("error while fetching user %s : %v ",userId, err)
		return nil, err
	}
	return user, nil 
}





// below Database methodds are to interact with a mysql database

var connectioUpgrader = websocket.Upgrader{
   ReadBufferSize: 1024,
   WriteBufferSize: 1024,
}



func GetAllUsers() ([]*models.User , error) {
	Db := GetSqlDb()
	if Db == nil {
		fmt.Println("db is nil")
		return nil, nil
	}

	result, err := Db.Query(`select id, username, name, email, mobile from users`)

	if err != nil {
		fmt.Println("error while fetching users in dbserivce")
		log.Fatal(err)
		return nil, err
	}

	fmt.Println("data fetched successfully")
	var users []*models.User
	for result.Next(){
		var user models.User
		err := result.Scan(&user.UserId, &user.UserName, &user.Name, &user.Email , &user.Contact)
		if err != nil {
			log.Fatal("error while scanning user")
			log.Fatal(err)
		}

		users = append(users, &user)
		log.Println(user)
	}


	return users, nil
}



func GetUser(id string) (*models.User, error)  {
	Db := GetSqlDb()
	if Db == nil {
		fmt.Println("db is nil")
		return nil, nil
	}

	result := Db.QueryRow(`select id, username, name, email, mobile from users where id = ?`, id)

	var user models.User
	if result != nil {
		err := result.Scan(&user.UserId, &user.UserName, &user.Name, &user.Email , &user.Contact)
		if err != nil {
			log.Println("error while scanning user : ", err)
			
			return nil, err
		}
	}else{
		log.Println("result is nul")
		return nil, errors.New("user not found")
	}

	return &user, nil
}	


func InsertNewUser(user *models.User) error {
	Db := GetSqlDb()
	if Db == nil {
		fmt.Println("db is nil")
		return errors.New("db connection not present")
	}

	_, err := Db.Exec(`insert into users (username, name, email, mobile) values (?, ?, ?, ?)`, user.UserName, user.Name, user.Email, user.Contact)
	if err != nil {
		log.Fatal("error while inserting user")
		log.Fatal(err)
		return err
	}

	return nil
}


func SendUsers(w http.ResponseWriter, r *http.Request){
	
	conn, err := connectioUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("error while upgrading connection")
		return
	}
	defer conn.Close()

	c := make(chan string)

	go func(){
		
		
		for {
			task, operating := <-c
			if !operating {	
				break
			}
			val , _ := strconv.Atoi(task)
			conn.WriteMessage(websocket.TextMessage, []byte(strconv.Itoa(val*2)))
		}
	}()

	msgType, msg, err := conn.ReadMessage()

	if err != nil {
		log.Println("error while reading message")
	}

	if msgType != websocket.TextMessage {
		conn.WriteMessage(msgType, []byte("message is not in text"))
		return 
	}

	if string(msg) != "start" {
		log.Println(string(msg))
		conn.WriteMessage(msgType,[]byte("message should be start"))
		return 
	}

	for {
		
		msgType, msg, err := conn.ReadMessage()	
		if err != nil {
			log.Println("error while reading message")
			close(c)
			return 	
		}

		if msgType != websocket.TextMessage {
			conn.WriteMessage(msgType, []byte("message is not in text"))
			return 
		}

		

		if string(msg) == "stop" {
			close(c)
			break	
		}

		c <- string(msg)
	}
}