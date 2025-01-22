package models

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2/table"
)


var MessageMetadata = table.Metadata{
	Name: "chat_store",
	Columns: []string{"user_id", "sender_id", "delivered", "received", "message", "event_id", "is_delivered"},
	PartKey: []string{"user_id", "sender_id"},
	SortKey: []string{"delivered"},
}

var MesageTable = table.New(MessageMetadata)

type Message struct {
	UserId string `json:"user_id" db:"user_id"`
	//MsgId      string `json:"msg_id"`
	SenderId   string `json:"sender_id" db:"sender_id"`
	Delivered time.Time `json:"delivered" db:"delivered"`
	Received  time.Time `json:"received" db:"received"`
	Message string `json:"message" db:"message"`
	EventId gocql.UUID `json:"event_id" db:"event_id"`
	IsDelivered bool `json:"is_delivered" db:"is_delivered"`

	
}

var UserMetadata = table.Metadata{
	Name: "user",
	Columns: []string{"user_id", "user_name", "name", "email", "contact", "password", "last_active"},
	PartKey: []string{"user_id"},
}

var UserTable = table.New(UserMetadata)

type User struct {
	UserId     string    `json:"user_id" db:"user_id"`
	UserName   string    `json:"user_name" db:"user_name"`
	Name       string    `json:"name" db:"name"`
	Email      string    `json:"email" db:"email"`
	Contact    string    `json:"contact" db:"contact"`
	Password   string    `json:"password" db:"password"`
	LastActive time.Time `json:"last_active" db:"last_active"`
}

type Friends struct {
	UserId    string   `json:"user_id"`
	FriendIds []string `json:"friend_ids"`
}

type Event struct {
	SenderId string `json:"sender_id"`
	Data   string `json:"data"`
	ReceiverId string `json:"receiver_id"`
}

type MsgEvent struct {
}