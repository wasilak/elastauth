# Elastauth (kibana-auth-proxy) example

Docker compose based example, couple of assumpions:

- you have a Elasticsearch cluster with Authentication & Authorization already enabled and working
- it was tested using podman, but should work with regular docker

Testing:

- run `docker-compose up`
- Traefik: <https://go-hello-world.127-0-0-1.nip.io:4343/>
- Nginx: <https://go-hello-world.127-0-0-1.nip.io:4444/>
- you should be redirected to Authelia
- You need to check in Elasticsearch/kibana if user was created
