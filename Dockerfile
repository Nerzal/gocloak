ARG KEYCLOAK_VERSION=26.8.2
FROM quay.io/keycloak/keycloak:${KEYCLOAK_VERSION}

ENV KC_BOOTSTRAP_ADMIN_USERNAME=admin
ENV KC_BOOTSTRAP_ADMIN_PASSWORD=secret
ENV KC_HOSTNAME=localhost
ENV KC_HEALTH_ENABLED=true
# tests use the v1 fine-grained admin permission endpoints (groups/clients .../management/permissions)
ENV KC_FEATURES=docker,scripts,admin-fine-grained-authz:v1

ENTRYPOINT ["/opt/keycloak/bin/kc.sh"]
CMD ["start-dev", "--import-realm"]
