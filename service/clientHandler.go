package service

import (
	"context"
	"dev/chatspace/models"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	conn   *websocket.Conn
	userId string
	buff   chan []byte
	manager *Manager
	redis_client *redis.Client
}


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}


func NewClient(conn *websocket.Conn,redis_client *redis.Client, userId string, manager *Manager) *Client {
	return &Client{
		conn: conn,
		userId : userId,
		buff: make(chan []byte),
		manager: manager,
		redis_client: redis_client,
	}
}


func (c *Client) register() {
	
	c.manager.register <- c
}

func (c *Client) receiveMessage(){
	c.conn.SetReadLimit(1024)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(payload string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		log.Println("pong received from user")
		return nil
	})
	// for client to receive messages and process it so it can be sent to the intended users
	for {
		msg_type , msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		if msg_type == websocket.TextMessage {
			log.Println("message received from client", string(msg))
			cleanedMsg := strings.ReplaceAll(strings.ReplaceAll(string(msg), "\r", "\\r"), "\n", "\\n")
			event := &models.Event{}
			log.Println("cleaned message : ", cleanedMsg)
			err := json.Unmarshal([]byte(cleanedMsg), event)
			if err != nil {
				log.Println("error occurred while unmarshalling event data : ", err)
				
			}else{
				c.manager.msgSent <- event
			}
		}	
	}
}

func (c *Client) sendMessage() {
	ctx := context.Background()
	client_sub := c.redis_client.Subscribe(ctx, c.userId)
	// for client to send messages to the inteded users
	defer close(c.buff)
	for {
		select {

			
			case msg := <-c.buff:
				err := c.conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println("error while sending message")
				}
				
			case rdb_msg := <-client_sub.Channel():
				err := c.conn.WriteMessage(websocket.TextMessage, []byte(rdb_msg.Payload))
				if err != nil {
					log.Println("error while sending message")
				}
		}
	}
}


func (c *Client) pingUser() {
	// for client to ping the user to check if the user is still active
	pinger := time.NewTicker(30 * time.Second)
	for range pinger.C {
			err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second))
			if err != nil {
			   log.Println("error while pinging user", err)
			   
			}
	}
}



func SocketHandler(manager *Manager, redis_client *redis.Client, w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	userId := r.URL.Query().Get("userId")
	createClient(manager, redis_client, socket, userId)
	log.Println("client connected : ", userId)
}

func createClient(manager *Manager, redis_client *redis.Client, socket *websocket.Conn, userId string) {
	client := NewClient(socket, redis_client, userId, manager )
	client.register()

	go client.receiveMessage()
	go client.sendMessage()
	go client.pingUser()

}