# A Real Time Chat System POC implementation based on the System Design by ALEX WU

## Overview
This project is a web application built using Golang, featuring the [Gochi](https://github.com/go-chi/chi) router for handling HTTP requests, [Gorilla WebSocket](https://github.com/gorilla/websocket) for real-time communication, [Cassandra](https://cassandra.apache.org/) as the primary database, and [Redis](https://redis.io/) for caching and session management.

## Features
- **Gochi Router**: Lightweight and fast HTTP routing.
- **Gorilla WebSocket**: Real-time, bidirectional communication.
- **Cassandra DB**: Scalable and distributed NoSQL database.
- **Redis**: High-performance in-memory key-value store for caching.

## Tech Stack
- **Language**: Golang
- **Web Framework**: Gochi
- **WebSocket**: Gorilla WebSocket
- **Database**: Cassandra
- **Pub Sub**: Redis Stream

## Reference - pretty helpful- 
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
   git clone https://github.com/yourusername/yourproject.git
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

## Usage
- Start the application and access it via `http://localhost:8080`
- WebSocket endpoint: `ws://localhost:8080/{user-id}/chat` - need to create a user and login with user to get jwt bearer token, which needs to be passed with endpoint



