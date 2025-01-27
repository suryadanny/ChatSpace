package service

import (
	"context"
	"dev/chatspace/dbservice"
	"dev/chatspace/models"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	conn   *websocket.Conn
	userId string
	buff   chan []byte
	closed chan bool
	manager *Manager
	redis_client *redis.Client
	repoStore *dbservice.RepoStore
}


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}


func NewClient(conn *websocket.Conn,redis_client *redis.Client, userId string, manager *Manager, repoStore *dbservice.RepoStore) *Client {
	return &Client{
		conn: conn,
		userId : userId,
		buff: make(chan []byte),
		closed: make(chan bool),
		manager: manager,
		redis_client: redis_client,
		repoStore: repoStore,
	}
}


func (c *Client) register() {
	
	c.manager.register <- c
}

func (c *Client) updateUserLastActiveTime(){
	c.repoStore.GetUserRepo().UpdateUser( map[string]interface{}{"last_active" : time.Now().UnixMilli()} , c.userId)
}

func (c *Client) receiveMessage(){

	defer close(c.closed)
	c.conn.SetReadLimit(1024)
	c.conn.SetReadDeadline(time.Now().Add(200 * time.Second))
	c.conn.SetPongHandler(func(payload string) error {
		c.conn.SetReadDeadline(time.Now().Add(200 * time.Second))
		c.updateUserLastActiveTime()
		//c.repoStore.GetUserRepo().UpdateUser( map[string]interface{}{"last_active" : time.Now().Unix()} , c.userId)
		//log.Println("pong received from user")
		return nil
	})
	// for client to receive messages and process it so it can be sent to the intended users
	for {
		msg_type , msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("error recieved during receving the message : ",err)
			if closeErr, ok := err.(*websocket.CloseError); ok {
				log.Println("close error received :", closeErr)
				c.manager.unregister <- c
				c.closed <- true
				return
			}
		}
		if msg_type == websocket.TextMessage {
			log.Println("message received from client", string(msg))
			cleanedMsg := strings.ReplaceAll(strings.ReplaceAll(string(msg), "\r", "\\r"), "\n", "\\n")
			event := &models.Event{}
			log.Println("cleaned message : ", cleanedMsg)
			err := json.Unmarshal([]byte(cleanedMsg), event)
			event.SenderId = c.userId
			event.IsDelivered = false
			event.Received = time.Now()
			
			event.EventId = gocql.TimeUUID()

			
			if err != nil {
				log.Println("error occurred while unmarshalling event data : ", err)
				
			}else{
				c.manager.msgSent <- event
				c.repoStore.GetEventRepo().AddEvent(event)
			}
		}	
	}
}

func (c *Client) sendMessage() {
	ctx := context.Background()
	client_sub := c.redis_client.Subscribe(ctx, c.userId)
	// for client to send messages to the inteded users
	close_chan := make(chan bool)
	defer close(c.buff)
	defer close(close_chan)
	go c.readFromStream(close_chan)
	

	for {
		select {

			
			case msg := <-c.buff:
				err := c.conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println("error while sending message")

				}else{
					event := &models.Event{}
					err := json.Unmarshal([]byte(msg), event)

					if err != nil {
						log.Println("error occurred while unmarshalling event data for updating event : ", err)
					}
					c.repoStore.GetEventRepo().UpdateEvent(map[string]interface{}{"is_delivered" : true, "delivered":time.Now().UnixMilli()}, event.EventId , event.SenderId, event.Received)
				}
				
			case rdb_msg := <-client_sub.Channel():
				log.Println("message received from redis : ", rdb_msg.Payload)
				err := c.conn.WriteMessage(websocket.TextMessage, []byte(rdb_msg.Payload))
				if err != nil {
					log.Println("error while sending message")
				}
			case <-c.closed:
				log.Println("client closed")
				close_chan <- true
				return
		}
	}
}

func (c *Client) readFromStream(close chan bool) {
	ctx := context.Background()
	id := "0"
	
	if c.repoStore.GetUserDeviceRepo().LastMsgIdRead(c.userId).RedisId != "" {
		id = c.repoStore.GetUserDeviceRepo().LastMsgIdRead(c.userId).RedisId
	}

	for {
		select {
			
			case <-close:
				log.Println("closing the stream")
				return
				
			default:

				msg, err  := c.redis_client.XRead(ctx, &redis.XReadArgs{
					Streams: []string{c.userId, id},
					Count: 2,
					Block: 1000 * time.Millisecond ,
				}).Result()


				if err != nil && err != redis.Nil {
					log.Println("error while reading from stream : ", err)
				}

				for _, stream := range msg {
					for _, message := range stream.Messages {
						log.Println("message received from redis : ", message)
						id = message.ID
						data := message.Values["payload"].(string)
						c.buff <- []byte(data)
					}
				}

				err = c.repoStore.GetUserDeviceRepo().UpdateUserDevice(map[string]interface{}{"redis_id" : id}, c.userId)
				if err != nil {
					log.Println("error while updating redis id : ", err)
				}
		}

	}
}


func (c *Client) pingUser() {
	// for client to ping the user to check if the user is still active
	pinger := time.NewTicker(60 * time.Second)
	for range pinger.C {
			err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second))
			if err != nil {
			   log.Println("error while pinging user", err)
			   if err == websocket.ErrCloseSent || err == websocket.ErrBadHandshake {
				 log.Println("close sent error received :", err)
			     return
			   }
			}
	}
}


func SocketHandler(manager *Manager, redis_client *redis.Client, w http.ResponseWriter, r *http.Request , repoStore *dbservice.RepoStore) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	userId := chi.URLParam(r, "id")
	//userId := r.URL.Query().Get("userId")
	createClient(manager, redis_client, socket, userId, repoStore)
	log.Println("client connected : ", userId)
}

func createClient(manager *Manager, redis_client *redis.Client, socket *websocket.Conn, userId string, repoStore *dbservice.RepoStore) {
	client := NewClient(socket, redis_client, userId, manager, repoStore)
	client.register()

	go client.receiveMessage()
	go client.sendMessage()
	go client.pingUser()

}