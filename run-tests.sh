#!/bin/sh

docker compose down
docker compose up -d

keycloakServer=http://localhost:8080
url="${keycloakServer}/health"
echo "Checking service availability at $url (CTRL+C to exit)"
while true; do
    response=$(curl -s -o /dev/null -w "%{http_code}" $url)
    if [ $response -eq 200 ]; then
        break
    fi
    sleep 1
done
echo "Service is now available at ${keycloakServer}"


go test 
docker compose down
