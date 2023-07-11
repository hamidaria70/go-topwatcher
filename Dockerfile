FROM golang:1.20.5-alpine

WORKDIR /app
COPY config.yaml /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /usr/local/bin/topwatcher

CMD [ "topwatcher" ]
