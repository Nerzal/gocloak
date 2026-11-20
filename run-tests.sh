#!/usr/bin/env bash
# usage: ./run-tests.sh [TestName ...]   (KEYCLOAK_VERSION=x.y.z to test another Keycloak release)

trap 'docker compose down' EXIT

docker compose down
docker compose up -d --build --wait --wait-timeout 300 || exit 1
echo "Keycloak is ready at http://localhost:8080"

ARGS=()
if [ $# -gt 0 ]; then
    IFS='|'
    ARGS+=("-run" "^($*)$")
    unset IFS
fi

go test -race -cover -coverprofile=coverage.out -covermode=atomic -p 10 -cpu 1,2 -bench . -benchmem "${ARGS[@]}" ./...
