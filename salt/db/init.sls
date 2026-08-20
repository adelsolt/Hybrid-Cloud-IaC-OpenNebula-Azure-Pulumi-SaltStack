postgresql:
  pkg.installed: []
  service.running:
    - enable: true
    - require:
      - pkg: postgresql

keycloak-user:
  postgres_user.present:
    - name: keycloak
    - password: {{ pillar['db_password'] }}
    - require:
      - service: postgresql

keycloak-db:
  postgres_database.present:
    - name: keycloak
    - owner: keycloak
    - require:
      - postgres_user: keycloak-user