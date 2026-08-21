#!jinja|yaml
remediate-config:
  local.state.apply:
    - tgt: {{ data['id'] }}
    - arg:
      - edge