#!/bin/sh

docker-compose up -d

go test -failfast -race -cover -coverprofile=coverage.txt -covermode=atomic -cpu 1,2 -bench . -benchmem

docker-compose down