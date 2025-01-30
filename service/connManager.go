package service

import (
	"context"
	"dev/chatspace/dbservice"
	"dev/chatspace/models"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

type Manager struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	msgSent    chan *models.Event
	redis_client *redis.Client
	close chan bool
	repoStore *dbservice.RepoStore
	// /lock 	 sync.RWMutex
}


// main manager that manages the clients and the messages that are sent to the clients
func NewManager(redis_client *redis.Client, repoStore *dbservice.RepoStore) *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		msgSent:    make(chan *models.Event),
		close: make(chan bool),
		redis_client: redis_client,
		repoStore: repoStore,
	}
}

//close the manager and all the channels that are present in the manager
func (m *Manager) Close(){
	m.close <- true
	close(m.close)
}

// start the manager and listen to the channels for the events
func (m *Manager) Start() {
	for {
		select {
		case client := <-m.register:
			// registering a new client
			m.clients[client.userId] = client	
			// add additional logic to check if messages are present in queue for them to be delivered
		case client := <-m.unregister:
			// unregistering a client
			log.Printf("client %s unregistered", client.userId)		
			delete(m.clients, client.userId)
		case event := <-m.msgSent:
			// a message is sent to the manager to be sent to a client
			log.Println("message received from client on manager : ", string(event.Message))
			
			log.Println("event data received from client on manager : ", string(event.Message))
			// extracts the receiver id and checks if the client is also conencted to the same server
			client, present := m.clients[event.ReceiverId]
			payload, err := json.Marshal(event)
			if err != nil {
				log.Println("error occurred while marshalling event data")
				continue
			}
		
			if present {
				// if the client is present in the same server then the message is direct sent to the respecitve client handler to be sent to the end user
				client.buff <- []byte(payload)

				log.Println("message sent to client",event.Message)
			} else {
				
				// if the client is not present in the same server then the message is added to the redis queue for the respective client
				// this will be consumed by the client itself on the other server subcribing to the redis queue by its user id
				entry_id, err:= m.redis_client.XAdd(context.Background(), &redis.XAddArgs{
					Stream: event.ReceiverId,
					Values: map[string]interface{}{"payload": payload},
				}).Result()
				// will be added to redis queue for users in another server

				if err != nil {
					log.Println("error occurred while publishing message to redis queue")
				}else{
					log.Println("message added to redis queue with entry id : ", entry_id)
				}
			}
		case <-m.close:
			//closing the manager
			log.Println("closing manager")
			return

		}
	}
}


// connManager only manages registration and unregistering and passing of clients to other clients inorder for them to communicates

