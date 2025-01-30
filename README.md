# A Real Time Chat System POC implementation based on - Building a real time chat system from System Design by ALEX WU

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
- WebSocket endpoint: `ws://localhost:8080/ws`

## Contributing
1. Fork the repository
2. Create a new branch (`git checkout -b feature-branch`)
3. Commit your changes (`git commit -m 'Add new feature'`)
4. Push to the branch (`git push origin feature-branch`)
5. Create a Pull Request

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact
For issues or feature requests, open an issue on GitHub or contact the maintainer.

