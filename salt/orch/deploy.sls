harden-all:
  salt.state:
    - tgt: '*'
    - sls: base

database:
  salt.state:
    - tgt: 'role:db'
    - tgt_type: grain
    - sls: db
    - require:
      - salt: harden-all

identity:
  salt.state:
    - tgt: 'role:app'
    - tgt_type: grain
    - sls: app
    - require:
      - salt: database

tunnel:
  salt.state:
    - tgt: 'role:app or role:edge'
    - tgt_type: compound
    - sls: wireguard
    - require:
      - salt: identity

edge-proxy:
  salt.state:
    - tgt: 'role:edge'
    - tgt_type: grain
    - sls: [edge, beacons]
    - require:
      - salt: tunnel