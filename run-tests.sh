#!/bin/sh

docker-compose down
docker-compose up -d

sleep 10

mkdir -p bin
go test -failfast -race -cover -coverprofile=bin/coverage.txt -covermode=atomic -cpu 1,2 -bench . -benchmem

docker-compose down