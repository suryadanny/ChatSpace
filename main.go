package main

import (
	"dev/chatspace/authentication"
	"dev/chatspace/dbservice"
	"dev/chatspace/service"
	"dev/chatspace/utils"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

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
	// /	fmt.Println("properties read successfully -- setting up db connections")
		dbservice.SetupSqlDbconnection(AppProperties)

	}


	//setting up the cassandra connection

	cluster := gocql.NewCluster(AppProperties["cql.hostname"])
	cluster.Keyspace = "store"
	cql_port, _ := strconv.Atoi(AppProperties["cql.port"])
	cluster.Port = cql_port
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	
	if err != nil {
		log.Fatal("error while creating session" ,err)
		return
	}
	// close the session when exiting the serivice
	defer session.Close()

	//creating cqlx session
	cqlx_session, err := gocqlx.WrapSession(session ,err)
	
	
	if err != nil {
		log.Fatal("error while wrapping cql session", err)
		return
	}


	//Container for all the repositories
	repoStore := dbservice.NewRepoStore()


	// instatiating user repository and setting it to the repoStore
	userRepo := dbservice.NewUserRepository(&cqlx_session)
	userDeviceRepo := dbservice.NewUserDeviceRepository(&cqlx_session)

	//setting the repositories to the repoStore
	repoStore.SetUserRepository(userRepo)
	repoStore.SetUserDeviceRepository(userDeviceRepo)
	repoStore.SetEventRepository(dbservice.NewEventRepository(&cqlx_session))
	
	//instantiating redis client to be used for 
	//communication between clients on multiple servers
	
	redis_addr := AppProperties["redis.hostname"] + ":" + AppProperties["redis.port"]
	

	redis_client := redis.NewClient(&redis.Options{
		Addr: redis_addr,
		Password: "",	
		DB: 0,
	})


	// a service hub in the application for moving messages betwee clients
	manager := service.NewManager(redis_client, repoStore)
    
	//instantiating user service
	userService := service.NewUserService(userRepo) 

	//starting the manager
	go manager.Start()

	defer manager.Close()

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	// public routes
	router.Group(func( r chi.Router){
		r.Post("/signup", utils.ValidateUserRequestMiddleWare(http.HandlerFunc(userService.CreateUser)))
		r.Post("/login", utils.ValidateLoginRequestMiddleWare(http.HandlerFunc(userService.Login)))
	})
     log.Println("server started at :", time.Now())

	//private routes with jwt auth tokens
	router.Group(func(r chi.Router){
		
		r.Route("/user/{id}", func(r chi.Router) {
		//jwt token authentication , this could be further explore to remove blacklisted tokens using bloom filter
		
			r.Use(authentication.TokenMiddleware)

		    r.Get("/", http.HandlerFunc(userService.GetUser)) 
			r.Get("/chat", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				service.SocketHandler(manager, redis_client, w, r ,repoStore)
			}))

			r.Get("/online/{userId}", http.HandlerFunc(userService.LastActive))
			r.Post("/delete", http.HandlerFunc(userService.DeleteUser))
			r.Put("/update", http.HandlerFunc(userService.UpdateUser))
		})

		r.Route("/system", func(r chi.Router) {
			r.Get("/AllUsers", http.HandlerFunc(userService.GetUsers))
		})
	})

	http.ListenAndServe(":"+port,router)
}