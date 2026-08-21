nginx:
  pkg.installed: []

edge-cert:
  cmd.run:
    - name: >
        openssl req -x509 -newkey rsa:2048 -nodes
        -keyout /etc/ssl/private/edge.key
        -out /etc/ssl/certs/edge.crt -days 365
        -subj "/CN={{ pillar['edge']['public_ip'] }}"
    - creates: /etc/ssl/certs/edge.crt

edge-vhost:
  file.managed:
    - name: /etc/nginx/sites-available/keycloak.conf
    - source: salt://edge/files/keycloak.conf
    - template: jinja

edge-vhost-enabled:
  file.symlink:
    - name: /etc/nginx/sites-enabled/keycloak.conf
    - target: /etc/nginx/sites-available/keycloak.conf
    - require:
      - file: edge-vhost

edge-default-off:
  file.absent:
    - name: /etc/nginx/sites-enabled/default

nginx-svc:
  service.running:
    - name: nginx
    - enable: true
    - watch:
      - file: edge-vhost
      - cmd: edge-cert
    - require:
      - file: edge-vhost-enabled