# ciscofon-server

A simple service that serves configuration files for Cisco IP phones over TFTP and HTTP.

It also exposes a live log of requests being made at `/dashboard`.


Example configuration:

```yaml
tftp:
  port: 69
  dir: /srv/tftp
http:
  port: 8080
  dir: /srv/http
```

# Docker usage

TODO
