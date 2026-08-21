wireguard:
  pkg.installed: []

wg-conf:
  file.managed:
    - name: /etc/wireguard/wg0.conf
    - source: salt://wireguard/files/wg0.conf.jinja
    - template: jinja
    - mode: 600

wg0:
  service.running:
    - name: wg-quick@wg0
    - enable: true
    - watch:
      - file: wg-conf

