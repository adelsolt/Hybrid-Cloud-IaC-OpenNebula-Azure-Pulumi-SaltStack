sshd-config:
  file.managed:
    - name: /etc/ssh/sshd_config
    - source: salt://base/files/sshd_config
    - mode: 644

sshd:
  service.running:
    - enable: true
    - watch:
      - file: sshd-config