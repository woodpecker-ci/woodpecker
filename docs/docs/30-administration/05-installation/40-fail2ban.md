# Fail2ban

Woodpecker logs every failed agent login together with the address of the client.
You can use [fail2ban](https://github.com/fail2ban/fail2ban) to block clients that try to guess agent tokens.

```json
{"level":"warn","error":"...","ip":"203.0.113.7","time":"2026-11-20T10:00:00Z","message":"agent authentication failed from 203.0.113.7"}
```

:::note
Agents that connect over a unix socket have no address and cannot be blocked.
:::

Keep the logs in JSON format, as only it contains the date fail2ban needs.
This is the default if the server does not run in a terminal, or set `WOODPECKER_DEBUG_PRETTY=false`.

## Filter

Use this filter if agents connect to the gRPC port of the server (`WOODPECKER_GRPC_ADDR`, default `:9000`) directly:

```ini title="/etc/fail2ban/filter.d/woodpecker-agent.conf"
[Definition]
failregex = agent authentication failed from <ADDR>(?:"|\s|$)
ignoreregex =
```

### Behind a proxy

Behind a proxy the server only sees the address of the proxy.
It then also logs the addresses it gets in the `X-Real-IP` and `X-Forwarded-For` headers:

```json
{"level":"warn","error":"...","ip":"10.0.0.2","x-forwarded-for":"198.51.100.1,203.0.113.7","time":"2026-11-20T10:00:00Z","message":"agent authentication failed from 10.0.0.2 x-forwarded-for=198.51.100.1,203.0.113.7"}
```

Clients can send these headers themselves, so only use them if agents can reach the gRPC port of the server through your proxy alone.
This filter takes the last address of `X-Forwarded-For`, the one your proxy added:

```ini title="/etc/fail2ban/filter.d/woodpecker-agent-proxy.conf"
[Init]
# header with the client address, set by your proxy: x-forwarded-for or x-real-ip
header = x-forwarded-for

[Definition]
failregex = agent authentication failed from \S+ (?:\S+ )*?<header>=(?:[^\s",]+,)*<ADDR>(?:"|\s|$)
ignoreregex =
```

Caddy and Traefik set `X-Forwarded-For` by default.
With nginx, add `grpc_set_header X-Forwarded-For $remote_addr;` next to `grpc_pass`.
If your proxy sets `X-Real-IP` instead and overwrites the value a client sent, use `filter = woodpecker-agent-proxy[header=x-real-ip]` in the jail.

The server does not understand the PROXY protocol, so use the headers instead.
Add the address of your proxy to `ignoreip` of the jail, so fail2ban never blocks it.

## Jail

Choose how fail2ban reads the logs of the server.
Behind a proxy, set `filter = woodpecker-agent-proxy` instead.

### systemd

If you use the [packages](./30-packages.md), the server logs to the journal:

```ini title="/etc/fail2ban/jail.d/woodpecker.local"
[woodpecker-agent]
enabled = true
filter = woodpecker-agent
backend = systemd
journalmatch = _SYSTEMD_UNIT=woodpecker-server.service
port = 9000
maxretry = 10
findtime = 10m
bantime = 1h
```

### Log file

Let the server write its logs to a file with `WOODPECKER_LOG_FILE=/var/log/woodpecker/server.log`:

```ini title="/etc/fail2ban/jail.d/woodpecker.local"
[woodpecker-agent]
enabled = true
filter = woodpecker-agent
logpath = /var/log/woodpecker/server.log
port = 9000
maxretry = 10
findtime = 10m
bantime = 1h
```

### Docker

Mount a folder of the host into the server container and use the log file setup above.
Docker passes published ports around the `INPUT` chain, so ban in the `DOCKER-USER` chain instead
and use the port inside the container:

```ini title="/etc/fail2ban/jail.d/woodpecker.local"
[woodpecker-agent]
enabled = true
filter = woodpecker-agent
logpath = /var/log/woodpecker/server.log
banaction = iptables-multiport
chain = DOCKER-USER
port = 9000
maxretry = 10
findtime = 10m
bantime = 1h
```

## Test

Check that the filter finds the failed logins in your logs:

```sh
fail2ban-regex /var/log/woodpecker/server.log /etc/fail2ban/filter.d/woodpecker-agent.conf
```

After a change of the jail, reload fail2ban with `fail2ban-client reload` and check it with `fail2ban-client status woodpecker-agent`.
