pyinotify:
  pip.installed:
    - bin_env: /opt/saltstack/salt/salt-pip

beacon-config:
  file.managed:
    - name: /etc/salt/minion.d/beacons.conf
    - contents: |
        beacons:
          inotify:
            - files:
                /etc/nginx/sites-enabled/keycloak.conf:
                  mask:
                    - modify
                    - delete
            - disable_during_state_run: True

restart-minion:
  cmd.run:
    - name: systemctl restart salt-minion
    - onchanges:
      - file: beacon-config