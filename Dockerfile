ARG KEYCLOAK_VERSION=26.8.2
FROM quay.io/keycloak/keycloak:${KEYCLOAK_VERSION}

ENV KC_BOOTSTRAP_ADMIN_USERNAME=admin
ENV KC_BOOTSTRAP_ADMIN_PASSWORD=secret
ENV KC_HOSTNAME=localhost
ENV KC_HEALTH_ENABLED=true
# preview: all preview features, e.g. legacy token exchange v1 (direct naked impersonation)
# admin-fine-grained-authz:v1: tests use the v1 management permission endpoints (users/groups/clients)
ENV KC_FEATURES=preview,docker,admin-fine-grained-authz:v1

ENTRYPOINT ["/opt/keycloak/bin/kc.sh"]
CMD ["start-dev", "--import-realm"]
