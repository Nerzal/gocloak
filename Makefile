test:
	./run-tests.sh

start-keycloak: stop-keycloak
	docker compose up -d --build

stop-keycloak:
	docker compose down

generate-gocloak-interface:
	@echo "Remember to: go install github.com/vburenin/ifacemaker@latest"
	@$(shell go env GOPATH)/bin/ifacemaker -f client.go -s GoCloak -i GoCloakIface -p gocloak -o gocloak_iface.go

# code-review-graph (optional dev tool, MIT): local code graph in .code-review-graph/. build = full parse (first time or
# broken graph); the git hooks and Claude hooks keep it current with update.
code-graph-build:
	code-review-graph build

# set up the AI coding tools found on this machine
code-graph-install:
	code-review-graph install

# what the last change touches
code-graph-changes:
	code-review-graph detect-changes --brief

# incremental update with the latest changes
code-graph-update:
	code-review-graph update
