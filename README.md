# A Real Time Chat System implementation based on the chapter from the Book System Design by ALEX WU

## Overview
A real time chat system built using Golang, featuring the [Gochi](https://github.com/go-chi/chi) router for handling HTTP requests, [Gorilla WebSocket](https://github.com/gorilla/websocket) for real-time communication, [Cassandra](https://cassandra.apache.org/) as the primary database, and [Redis](https://redis.io/) for caching and session management.

## Features
- **Gochi Router**: Lightweight and fast HTTP routing.
- **Gorilla WebSocket**: Real-time, bidirectional communication.
- **Cassandra DB**: Scalable and distributed NoSQL database.
- **Redis**: For streaming of events/message between clients on different servers.

## Tech Stack
- **Language**: Golang
- **Web Framework**: Gochi
- **WebSocket**: Gorilla WebSocket
- **Database**: Cassandra
- **Pub Sub**: Redis Stream

## References - pretty helpful 
- https://timothy-urista.medium.com/an-easy-guide-to-implementing-pagination-in-cassandra-using-go-e7d13cfc804a
- https://bitek.dev/blog/go_cassandra_gocql/
- https://codesahara.com/blog/golang-job-queue-with-redis-streams/ - better understanding of redis streams
- https://github.com/scylladb/gocqlx/blob/master/example_test.go - for gocqlx usage
- https://redis.uptrace.dev/guide/go-redis-pubsub.html - redis pub/sub 
- https://redis.io/docs/latest/develop/data-types/streams/ - redis stream implementation
- https://antirez.com/news/114 - great read on redis streams
- ConnManager Implementation referenced from - https://github.com/gorilla/websocket/tree/main/examples/chat

## Installation

### Prerequisites
- [Golang](https://go.dev/dl/) installed
- [Cassandra](https://cassandra.apache.org/download/) running
- [Redis](https://redis.io/download/) running

### Steps
1. Clone the repository:
   ```sh
   git clone https://github.com/suryadanny/ChatSpace
   cd yourproject
   ```
2. Install dependencies:
   ```sh
   go mod tidy
   ```

3. Run the application:
   ```sh
   go run main.go <port>
   ```

## Project Structure
```
├── main.go
├── authentication/
│   ├── AuthenticationService.go
├── dbservice/
│   ├── dbservice.go
│   ├── eventRepo.go
│   ├── UserDeviceRepo.go
│   ├── userRepo.go
├── models/
│   ├── mdoel.go
├── service/
│   ├── clientHandler.go
│   ├── connManager.go
│   ├── userService.go
├── utils/
│   ├── constants.go
│   ├── properties.go
│   ├── validations.go
├── app.properties
├── go.mod
├── go.sum
└── README.md
```

## Usage & Endpoints
- Start the application and access it via `http://localhost:8080`
- WebSocket endpoint: `ws://localhost:8080/{user-id}/chat` - need to create a user and login with user to get jwt bearer token, which needs to be passed with endpoint

## API Specification

#### Signup Endpoint
##### URL:
`POST http://localhost:8000/signup`

##### Request Body:
```json
{
  "user_name": "******",
  "name": "******",
  "email": "******@******.com",
  "contact": "**********",
  "password" : "******"
}
```

#### Login Endpoint
##### URL:
`POST http://localhost:8000/login`

##### Request Body:
```json
{
  "user_name": "******",
  "password": "******"
}
```

#### Update User Endpoint
##### URL:
`PUT http://localhost:8000/user/{user_id}/update`

##### Request Body (example for updating password):
```json
{
  "password" : "******"
}
```

#### WebSocket Chat Endpoint
##### URL:
`ws://localhost:8000/user/{user_id}/chat`

##### Headers:
- **Authorization**: Bearer token

#### Get Friend's Last Online Status
##### URL:
`GET http://localhost:8000/user/{user_id}/online/{friend_id}`

##### Response:
```json
{
  "friend_id": "******",
  "last_online": "YYYY-MM-DDTHH:MM:SSZ"
}
```




