test:
	./run-tests.sh

start-keycloak: stop-keycloak
	docker compose up -d

stop-keycloak:
	docker compose down

start-keycloak-old: stop-keycloak
	docker-compose up -d

stop-keycloak-old:
	docker-compose down