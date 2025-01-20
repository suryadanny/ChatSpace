package main

import (
	"dev/chatspace/dbservice"
	"dev/chatspace/service"
	"dev/chatspace/utils"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"
	"github.com/scylladb/gocqlx/v2"
)




func main(){
	port := "8000"
	
	if len(os.Args) < 2 {
		log.Println("Port not provided, using default port 8000")
		
	}else{
		port = os.Args[1]
	}
	
	log.Println("Starting the server on port : ", port)

	AppProperties, err := utils.GetAppPropeties()
	if err != nil {
		log.Fatal("error while reading properties")
		return
	}else{
		fmt.Println("properties read successfully -- setting up db connections")
		dbservice.SetupSqlDbconnection(AppProperties)

	}
	fmt.Println(AppProperties)

	cluster := gocql.NewCluster(AppProperties["cqsql.hostname"])
	cluster.Keyspace = "store"
	cluster.Port = 9042
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	
	if err != nil {
		log.Fatal("error while creating session" ,err)
		return
	}
	// close the session when exiting the serivice
	defer session.Close()

	//creating user respository
	cqlx_session, err := gocqlx.WrapSession(session ,err)
	
	
	if err != nil {
		log.Fatal("error while wrapping cql session", err)
		return
	}

	userRepo := dbservice.NewUserRepository(&cqlx_session)
	//instantiating redis client to be used for 
	//communication between clients on multiple servers
	redis_client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",	
		DB: 0,
	})
	// a service hub in the application for moving messages betwee clients
	manager := service.NewManager(redis_client)
    userService := service.NewUserService(userRepo) 

	go manager.Start()

	
	
	
	


	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	

	router.Route("/user", func(r chi.Router) {


		r.Get("/", http.HandlerFunc(userService.GetUsers))
		r.Get("/{id}", http.HandlerFunc(userService.GetUser)) 
		r.Post("/", http.HandlerFunc(userService.CreateUser))
		r.Get("/{id}/chat", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			service.SocketHandler(manager, redis_client, w, r)
		}))
		r.Post("/{id}/delete", http.HandlerFunc(userService.DeleteUser))
		r.Put("/{id}/update", http.HandlerFunc(userService.UpdateUser))
	})
	

	// router.Get("/chat", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	service.SocketHandler(manager, redis_client, w, r)
	// }))


	http.ListenAndServe(":"+port,router)
}