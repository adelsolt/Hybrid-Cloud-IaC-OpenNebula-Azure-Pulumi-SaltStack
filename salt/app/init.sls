openjdk:
  pkg.installed:
    - name: openjdk-21-jre-headless

keycloak-user:
  user.present:
    - name: keycloak
    - system: true
    - shell: /usr/sbin/nologin

keycloak-tarball:
  archive.extracted:
    - name: /opt/keycloak
    - source: https://github.com/keycloak/keycloak/releases/download/26.0.0/keycloak-26.0.0.tar.gz
    - source_hash: sha256=<paste-real-hash>
    - user: keycloak
    - options: --strip-components=1
    - enforce_toplevel: false

keycloak-conf:
  file.managed:
    - name: /opt/keycloak/conf/keycloak.conf
    - source: salt://app/files/keycloak.conf
    - template: jinja

keycloak-unit:
  file.managed:
    - name: /etc/systemd/system/keycloak.service
    - source: salt://app/files/keycloak.service

keycloak:
  service.running:
    - enable: true
    - watch:
      - file: keycloak-conf
      - file: keycloak-unit
    - require:
      - archive: keycloak-tarball