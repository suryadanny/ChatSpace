package main

import (
	"dev/chatspace/dbservice"
	"dev/chatspace/service"
	"dev/chatspace/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/redis/go-redis/v9"

	// "github.com/go-sql-driver/mysql"
	//
	_ "github.com/go-sql-driver/mysql"
)




func main(){
	
	AppProperties, err := utils.GetAppPropeties()
	if err != nil {
		log.Fatal("error while reading properties")
		return
	}else{
		fmt.Println("properties read successfully -- setting up db connections")
		dbservice.SetupSqlDbconnection(AppProperties)

	}
	fmt.Println(AppProperties)

	redis_client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",	
		DB: 0,
	})
	// a service hub in the application for moving messages betwee clients
	manager := service.NewManager(redis_client)
    userService := service.NewUserService() 

	go manager.Start()

	
	
	
	


	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	

	router.Route("/", func(r chi.Router) {
		r.Get("/user", http.HandlerFunc(userService.GetUsers))
		r.Get("/user/{id}", http.HandlerFunc(userService.GetUser)) 
		r.Post("/user", http.HandlerFunc(userService.CreateUser))

		// r.Route("/chat", func(r chi.Router) {
		// 	r.Ge
		// })
	})
	

	router.Get("/websocket", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service.SocketHandler(manager, redis_client, w, r)
	}))


	http.ListenAndServe(":8000",router)
}