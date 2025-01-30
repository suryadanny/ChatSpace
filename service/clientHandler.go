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

//the below code, handle socket connection for the client, and the responsibility of receiving and delivering the message to the intended user


//creates a new client with the connection, redis client, user id, manager and repo store
func NewClient(conn *websocket.Conn,redis_client *redis.Client, userId string, manager *Manager, repoStore *dbservice.RepoStore) *Client {
	return &Client{
		conn: conn,
		userId : userId,
		//chnnel to receive the message from other clients
		buff: make(chan []byte),
		//channel to close the client and also used to close goroutines of the client handler
		closed: make(chan bool),
		manager: manager,
		redis_client: redis_client,
		repoStore: repoStore,
	}
}


//register the client to the manager
func (c *Client) register() {
	
	c.manager.register <- c
}

//updating the last active time of the user, so show if the user if online or not to other friends of the user
func (c *Client) updateUserLastActiveTime(){
	c.repoStore.GetUserRepo().UpdateUser( map[string]interface{}{"last_active" : time.Now().UnixMilli()} , c.userId)
}

//receive message from the client and run as a goroutine
func (c *Client) receiveMessage(){

	defer close(c.closed)
	c.conn.SetReadLimit(1024)
	c.conn.SetReadDeadline(time.Now().Add(200 * time.Second))
	c.conn.SetPongHandler(func(payload string) error {
		//increase the read deadline of the connection, to keep the connection alive
		c.conn.SetReadDeadline(time.Now().Add(200 * time.Second))
		//update the last active time of the user
		c.updateUserLastActiveTime()
		return nil
	})
	// for client to receive messages and process it so it can be sent to the intended users
	for {
		//read the message from the user's socket connection
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
			//clean the message and unmarshal the message to the event struct
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
				//send the message to the manager to be sent to the intended user, if the user is connected to the same server then the message is directly sent to the client handler
				//or else passed to the redis stream to be sent to the user, to topic usually named as the receiver user id
				c.manager.msgSent <- event
				//add the event to the event store in db
				c.repoStore.GetEventRepo().AddEvent(event)
			}
		}	
	}
}


//send message to the client and run as a goroutine
func (c *Client) sendMessage() {
	ctx := context.Background()
	client_sub := c.redis_client.Subscribe(ctx, c.userId)
	// for client to send messages to the inteded users
	close_chan := make(chan bool)
	defer close(c.buff)
	defer close(close_chan)

	//subsribe to redis stream of topic name as its own user_id , and read from the redis stream for the user to
	// check if there are any messages for the user from other user on other servers
	//starting a goroutine to read from the redis stream
	go c.readFromStream(close_chan)
	
	//select and channels make our execution thred safe and also helps in managing the goroutines
	for {
		select {

			//the buff channel is both used by the conn manager , and redis goroutine to send the message to the client in a thread safe manner
			case msg := <-c.buff:
				//we write message to websocket connection
				err := c.conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println("error while sending message")

				}else{
					event := &models.Event{}
					err := json.Unmarshal([]byte(msg), event)

					if err != nil {
						log.Println("error occurred while unmarshalling event data for updating event : ", err)
					}
					//update the event/message as delivered in the event store
					c.repoStore.GetEventRepo().UpdateEvent(map[string]interface{}{"is_delivered" : true, "delivered":time.Now().UnixMilli()}, event.EventId , event.SenderId, event.Received)
				}
			case rdb_msg := <-client_sub.Channel():
				// we can use this if we plan to use redis pub/sub instead of stream, but stream is more reliable and can be used to store the messages for the user	
				log.Println("message received from redis : ", rdb_msg.Payload)
				err := c.conn.WriteMessage(websocket.TextMessage, []byte(rdb_msg.Payload))
				if err != nil {
					log.Println("error while sending message")
				}
			case <-c.closed:
				//close the connection and exit the goroutines
				log.Println("client closed")
				// close channel to stop the goroutine listening to the redis stream
				close_chan <- true
				return
		}
	}
}

//read from the redis stream for the user
func (c *Client) readFromStream(close chan bool) {
	ctx := context.Background()
	id := "0"
	// getting the offset message id from the user device table, to used to read messages from that offset
	if c.repoStore.GetUserDeviceRepo().LastMsgIdRead(c.userId).RedisId != "" {
		id = c.repoStore.GetUserDeviceRepo().LastMsgIdRead(c.userId).RedisId
	}

	for {
		select {
			
			case <-close:
				log.Println("closing the stream")
				return
				
			default:
				//read from the redis stream, its blocking and will wait for the message to be published to the stream, to avoid excessive polling
				msg, err  := c.redis_client.XRead(ctx, &redis.XReadArgs{
					Streams: []string{c.userId, id},
					Count: 2,
					Block: 1000 * time.Millisecond ,
				}).Result()


				if err != nil && err != redis.Nil {
					log.Println("error while reading from stream : ", err)
				}
				//parsing the stream message and sending it to the client buff channels to be written to the websocket connection
				for _, stream := range msg {
					for _, message := range stream.Messages {
						log.Println("message received from redis : ", message)
						id = message.ID
						data := message.Values["payload"].(string)
						c.buff <- []byte(data)
					}
				}
				//update the offset message id in the user device table
				err = c.repoStore.GetUserDeviceRepo().UpdateUserDevice(map[string]interface{}{"redis_id" : id}, c.userId)
				if err != nil {
					log.Println("error while updating redis id : ", err)
				}
		}

	}
}


func (c *Client) pingUser() {
	// for client to ping the user to check if the user is still active, pinging user every 60 seconds
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

//handler for the socket connection, to upgrade the https connection to websocket
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
	//creates the client , registers its and starts the below goroutines
	go client.receiveMessage()
	go client.sendMessage()
	go client.pingUser()

}