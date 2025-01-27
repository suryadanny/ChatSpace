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


func (m *Manager) Close(){
	m.close <- true
	close(m.close)
}

func (m *Manager) Start() {
	for {
		select {
		case client := <-m.register:
			// though select is thread safe, placing a lock if the hub
			m.clients[client.userId] = client	
			// add additional logic to check if messages are present in queue for them to be delivered
		case client := <-m.unregister:
			log.Printf("client %s unregistered", client.userId)		
			delete(m.clients, client.userId)
		case event := <-m.msgSent:
			log.Println("message received from client on manager : ", string(event.Message))
			
			log.Println("event data received from client on manager : ", string(event.Message))
			client, present := m.clients[event.ReceiverId]
			payload, err := json.Marshal(event)
			if err != nil {
				log.Println("error occurred while marshalling event data")
				continue
			}
		
			if present {
				client.buff <- []byte(payload)

				log.Println("message sent to client",event.Message)
			} else {
				
				// /err := m.redis_client.Publish(context.Background(), event.UserId, event.Data).Err()
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
			log.Println("closing manager")
			return

		}
	}
}


// hub only manages registration and unregistering and passing of clients to other clients inorder for them to communicates

// flows needed to handle if the messages are present in the redis queue it has to conumed and this can be directly done by the client itself, the 