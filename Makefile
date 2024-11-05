test:
	./run-tests.sh

start-keycloak: stop-keycloak
	docker compose up -d

stop-keycloak:
	docker compose down

generate-gocloak-interface:
	@echo "Remember to: go install github.com/vburenin/ifacemaker@latest"
	@$(shell go env GOPATH)/bin/ifacemaker -f client.go -s GoCloak -i GoCloakIface -p gocloak -o gocloak_iface.go
