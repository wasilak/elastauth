# elastauth

[![Docker Repository on Quay](https://quay.io/repository/wasilak/elastauth/status "Docker Repository on Quay")](https://quay.io/repository/wasilak/elastauth) [![CI](https://github.com/wasilak/elastauth/actions/workflows/main.yml/badge.svg)](https://github.com/wasilak/elastauth/actions/workflows/main.yml) [![Maintainability](https://api.codeclimate.com/v1/badges/d75cc6b44c7c33f0b530/maintainability)](https://codeclimate.com/github/wasilak/elastauth/maintainability) [![Go Reference](https://pkg.go.dev/badge/github.com/wasilak/elastauth.svg)](https://pkg.go.dev/github.com/wasilak/elastauth)

<img align="left" src="https://github.com/wasilak/elastauth/blob/main/gopher.png?raw=true" width="40%" height="40%" />

Designed to work as a forwardAuth proxy for Traefik (possibly others, like nginx, but not tested) in order to use LDAP/Active Directory for user access in Elasticsearch without paid subscription.

1. Request goes to Traefik
2. Traefik proxies it to Authelia in order to verify user
3. If it receives `200` forwards headers from Authelia to second auth -> kibana-auth-proxy
4. kibana-proxy-auth:
   - generates random password for local Kibana user (has nothing to do with LDAP password)
   - uses information from Authelia headers to create/update local user in Kibana + AD group/kibana roles mappings from config file
   - generates and passes back to Traefik header:

      ```
      Authorization: Basic XXXYYYZZZZ
      ```

5. Traefik passes user to Kibana with `Authorization` header which has password already set by kibana-proxy-pass and logs him/her in :)
6. Passwords are meant to have short time span of life and are regenerated transparently for user while using Kibana

Headers used by Authelia and kibana-auth-proxy:

```
remote-email
remote-groups
remote-name
remote-user
```

![architecture](https://github.com/wasilak/kibana-auth-proxy/blob/main/kibana-auth-proxy.png?raw=true)
