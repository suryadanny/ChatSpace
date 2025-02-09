FROM golang:1.22 as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY app.properties ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /chatspace

EXPOSE 8000

CMD [ "/chatspace" ]
