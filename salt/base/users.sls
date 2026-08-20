{% for name, u in pillar.get('users', {}).items() %}
user-{{ name }}:
  user.present:
    - name: {{ name }}
    - shell: /bin/bash
    - groups:
      - sudo

key-{{ name }}:
  ssh_auth.present:
    - user: {{ name }}
    - name: {{ u['key'] }}
    - require:
      - user: user-{{ name }}
{% endfor %}