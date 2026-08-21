postgresql:
  pkg.installed: []
  service.running:
    - enable: true
    - require:
      - pkg: postgresql
    - watch:
      - file: pg-listen
      - file: pg-hba

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

pg-listen:
  file.replace:
    - name: /etc/postgresql/14/main/postgresql.conf
    - pattern: "^#?listen_addresses.*"
    - repl: "listen_addresses = '127.0.0.1,10.0.0.233'"
    - append_if_not_found: true

pg-hba:
  file.append:
    - name: /etc/postgresql/14/main/pg_hba.conf
    - text: "host keycloak keycloak 10.0.0.234/32 scram-sha-256"