nftables:
  pkg.installed: []

nft-ruleset:
  file.managed:
    - name: /etc/nftables.conf
    - source: salt://base/files/nftables.conf
    - mode: 644

nftables-svc:
  service.running:
    - name: nftables
    - enable: true
    - watch:
      - file: nft-ruleset