package models

type Message struct {
	Content string `json:"content"`
	//MsgId      string `json:"msg_id"`
	SenderId   string `json:"sender_id"`
	ReceiverId string `json:"receiver_id"`
	CreatedAt  string `json:"created_at"`
	GroupId    string `json:"group_id"`
}

type User struct {
	UserId   string `json:"user_id"`
	UserName string `json:"user_name"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Contact  string `json:"contact"`
}

type Friends struct {
	UserId    string   `json:"user_id"`
	FriendIds []string `json:"friend_ids"`
}

type Event struct {
	UserId string `json:"user_id"`
	Data   string `json:"data"`
}