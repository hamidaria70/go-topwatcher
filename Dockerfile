# Build the application from source
FROM golang:1.20.5-alpine AS build-stage

WORKDIR /app

COPY . /app
RUN go mod download

RUN go build -o /topwatcher

# Deploy the application binary into a lean image
FROM alpine:latest AS build-release-stage

WORKDIR /app_topwatcher

COPY . /app_topwatcher
COPY --from=build-stage /topwatcher /usr/local/bin

CMD [ "topwatcher" ]
