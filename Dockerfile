FROM quay.io/keycloak/keycloak:25.0.1
COPY testdata data/import
WORKDIR /opt/keycloak
ENV KC_HOSTNAME=localhost
ENV KEYCLOAK_USER=admin
ENV KEYCLOAK_PASSWORD=secret
ENV KEYCLOAK_ADMIN=admin
ENV KEYCLOAK_ADMIN_PASSWORD=secret
ENV KC_HEALTH_ENABLED=true
ENV KC_FEATURES=account-api,account3,authorization,client-policies,impersonation,docker,scripts,admin-fine-grained-authz
RUN /opt/keycloak/bin/kc.sh import --file /data/import/gocloak-realm.json
ENTRYPOINT ["/opt/keycloak/bin/kc.sh"]
