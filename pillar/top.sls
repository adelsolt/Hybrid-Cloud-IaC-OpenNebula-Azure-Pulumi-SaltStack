base:
  '*':
    - common
  'role:db':
    - match: grain
    - db
    - secrets

  'role:app':
    - match: grain
    - app
    - secrets