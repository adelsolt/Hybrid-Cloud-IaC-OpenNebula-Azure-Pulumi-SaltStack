base:
  '*':
    - base
  'role:db':
    - match: grain
    - db
  'role:app':
    - match: grain
    - app
    - wireguard
  'role:edge':
    - match: grain
    - edge
    - wireguard