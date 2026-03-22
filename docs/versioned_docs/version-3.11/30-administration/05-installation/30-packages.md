# Binary Installation Guide

This guide explains how to install and configure Woodpecker CI using binary packages without Docker. This is useful for scenarios where you prefer a more traditional deployment method or need tighter integration with the host system.

## Prerequisites

Before installing Woodpecker, ensure your system meets the following requirements:

- **Operating System**: Linux (amd64, arm64, or armv7)
- **Init System**: systemd (for service management)
- **Database**: SQLite (default, embedded) or MySQL/PostgreSQL (recommended for production)
- **Git Forge**: Access to a Git forge (GitHub, GitLab, Gitea, etc.) for OAuth app setup
- **Network**: Server must be accessible from the internet (for webhooks) or your internal network

## Installation

### Step 1: Download Packages

Download the latest release packages from GitHub:

```bash
# Get latest release version
RELEASE_VERSION=$(curl -s https://api.github.com/repos/woodpecker-ci/woodpecker/releases/latest | grep -Po '"tag_name":\s"v\K[^"]+')

# For Debian/Ubuntu (x86_64)
curl -fLO "https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-server_${RELEASE_VERSION}_amd64.deb"
curl -fLO "https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-agent_${RELEASE_VERSION}_amd64.deb"
curl -fLO "https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-cli_${RELEASE_VERSION}_amd64.deb"

# For CentOS/RHEL/Rocky Linux (x86_64)
curl -fLO "https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-server-${RELEASE_VERSION}.x86_64.rpm"
curl -fLO "https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-agent-${RELEASE_VERSION}.x86_64.rpm"
curl -fLO "https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-cli-${RELEASE_VERSION}.x86_64.rpm"
```

### Step 2: Install Packages

Install the downloaded packages:

```bash
# Debian/Ubuntu
sudo apt --fix-broken install ./woodpecker-server_${RELEASE_VERSION}_amd64.deb
sudo apt --fix-broken install ./woodpecker-agent_${RELEASE_VERSION}_amd64.deb
sudo apt --fix-broken install ./woodpecker-cli_${RELEASE_VERSION}_amd64.deb

# CentOS/RHEL/Rocky Linux
sudo dnf install ./woodpecker-server-${RELEASE_VERSION}.x86_64.rpm
sudo dnf install ./woodpecker-agent-${RELEASE_VERSION}.x86_64.rpm
sudo dnf install ./woodpecker-cli-${RELEASE_VERSION}.x86_64.rpm
```

The installation will:
- Create a `woodpecker` user and group
- Install binaries to `/usr/local/bin/`
- Create systemd service files
- Create configuration directories
- Set up log directories

### Step 3: Configure the Server

Woodpecker requires a Git forge OAuth application. This example uses GitHub:

1. Go to GitHub → Settings → Developer settings → OAuth Apps → New OAuth App
2. Set Authorization callback URL to: `https://your-domain.com/authorize`
3. Note the Client ID and Client Secret

Create the server environment file:

```bash
sudo cp /etc/woodpecker/woodpecker-server.env.example /etc/woodpecker/woodpecker-server.env
sudo chmod 600 /etc/woodpecker/woodpecker-server.env
```

Edit `/etc/woodpecker/woodpecker-server.env`:

```ini
# Server configuration
WOODPECKER_HOST=https://ci.example.com
WOODPECKER_SERVER_ADDR=:8000
WOODPECKER_GRPC_ADDR=:9000

# GitHub OAuth
WOODPECKER_GITHUB=true
WOODPECKER_GITHUB_CLIENT=your-github-client-id
WOODPECKER_GITHUB_SECRET=your-github-secret

# Security
WOODPECKER_AGENT_SECRET=$(openssl rand -hex 32)
WOODPECKER_OPEN=true

# Database (SQLite - default)
WOODPECKER_DATABASE_DRIVER=sqlite3
WOODPECKER_DATABASE_DATASOURCE=/var/lib/woodpecker/woodpecker.sqlite

# Or use MySQL
# WOODPECKER_DATABASE_DRIVER=mysql
# WOODPECKER_DATABASE_DATASOURCE=woodpecker:password@tcp(localhost:3306)/woodpecker?parseTime=true

# Or use PostgreSQL
# WOODPECKER_DATABASE_DRIVER=postgres
# WOODPECKER_DATABASE_DATASOURCE=postgres://woodpecker:password@localhost:5432/woodpecker?sslmode=disable
```

### Step 4: Configure the Agent

Create the agent environment file:

```bash
sudo cp /etc/woodpecker/woodpecker-agent.env.example /etc/woodpecker/woodpecker-agent.env
sudo chmod 600 /etc/woodpecker/woodpecker-agent.env
```

Edit `/etc/woodpecker/woodpecker-agent.env`:

```ini
# Server connection
WOODPECKER_SERVER=localhost:9000
WOODPECKER_AGENT_SECRET=match-server-secret-from-above

# Agent capabilities
WOODPECKER_MAX_WORKFLOWS=4
WOODPECKER_BACKEND=docker

# Docker socket (for Docker backend)
DOCKER_HOST=unix:///var/run/docker.sock
```

### Step 5: Database Setup (Optional)

If using MySQL or PostgreSQL instead of SQLite:

**MySQL:**
```sql
CREATE DATABASE woodpecker CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'woodpecker'@'localhost' IDENTIFIED BY 'strong-password';
GRANT ALL PRIVILEGES ON woodpecker.* TO 'woodpecker'@'localhost';
FLUSH PRIVILEGES;
```

**PostgreSQL:**
```sql
CREATE USER woodpecker WITH PASSWORD 'strong-password';
CREATE DATABASE woodpecker OWNER woodpecker;
GRANT ALL PRIVILEGES ON DATABASE woodpecker TO woodpecker;
```

### Step 6: Start Services

Enable and start the services:

```bash
# Reload systemd to pick up new services
sudo systemctl daemon-reload

# Enable services to start on boot
sudo systemctl enable woodpecker-server
sudo systemctl enable woodpecker-agent

# Start services
sudo systemctl start woodpecker-server
sudo systemctl start woodpecker-agent

# Check status
sudo systemctl status woodpecker-server
sudo systemctl status woodpecker-agent
```

### Step 7: Verify Installation

Check that the services are running:

```bash
# View logs
sudo journalctl -u woodpecker-server -f
sudo journalctl -u woodpecker-agent -f

# Test CLI
woodpecker-cli info

# Access web UI
# Open https://ci.example.com in your browser
```

## Systemd Service Details

The packages install these systemd service files:

### Server Service

```ini
# /usr/local/lib/systemd/system/woodpecker-server.service
[Unit]
Description=Woodpecker CI Server
Documentation=https://woodpecker-ci.org/docs/administration/server-config
Requires=network.target
After=network.target
ConditionFileNotEmpty=/etc/woodpecker/woodpecker-server.env

[Service]
Type=simple
EnvironmentFile=/etc/woodpecker/woodpecker-server.env
User=woodpecker
Group=woodpecker
ExecStart=/usr/local/bin/woodpecker-server
WorkingDirectory=/var/lib/woodpecker/
StateDirectory=woodpecker
Restart=on-failure
RestartSec=10

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/woodpecker

[Install]
WantedBy=multi-user.target
```

### Agent Service

```ini
# /usr/local/lib/systemd/system/woodpecker-agent.service
[Unit]
Description=Woodpecker CI Agent
Documentation=https://woodpecker-ci.org/docs/administration/agent-config
Requires=network.target
After=network.target woodpecker-server.service
ConditionFileNotEmpty=/etc/woodpecker/woodpecker-agent.env

[Service]
Type=simple
EnvironmentFile=/etc/woodpecker/woodpecker-agent.env
User=woodpecker
Group=woodpecker
ExecStart=/usr/local/bin/woodpecker-agent
WorkingDirectory=/var/lib/woodpecker/
Restart=on-failure
RestartSec=10

# For Docker backend, add woodpecker user to docker group
# usermod -aG docker woodpecker

[Install]
WantedBy=multi-user.target
```

## Reverse Proxy Setup

### Nginx

```nginx
server {
    listen 80;
    server_name ci.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name ci.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Apache

```apache
<VirtualHost *:80>
    ServerName ci.example.com
    Redirect permanent / https://ci.example.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName ci.example.com

    SSLEngine on
    SSLCertificateFile /path/to/cert.pem
    SSLCertificateKeyFile /path/to/key.pem

    ProxyPreserveHost On
    ProxyPass / http://localhost:8000/
    ProxyPassReverse / http://localhost:8000/

    # WebSocket support
    RewriteEngine on
    RewriteCond %{HTTP:Upgrade} websocket [NC]
    RewriteRule ^/?(.*) "ws://localhost:8000/$1" [P,L]
</VirtualHost>
```

## Troubleshooting

### Service fails to start

```bash
# Check for configuration errors
sudo woodpecker-server --help
sudo -u woodpecker /usr/local/bin/woodpecker-server

# View detailed logs
sudo journalctl -u woodpecker-server -n 100 --no-pager
```

### Agent cannot connect to server

```bash
# Verify network connectivity
telnet localhost 9000

# Check firewall rules
sudo iptables -L -n | grep 9000

# Verify agent secret matches server
sudo cat /etc/woodpecker/woodpecker-server.env | grep AGENT_SECRET
sudo cat /etc/woodpecker/woodpecker-agent.env | grep AGENT_SECRET
```

### Permission denied errors

```bash
# Fix ownership
sudo chown -R woodpecker:woodpecker /var/lib/woodpecker
sudo chown -R woodpecker:woodpecker /etc/woodpecker
sudo chmod 600 /etc/woodpecker/*.env

# For Docker backend
sudo usermod -aG docker woodpecker
```

### Database connection issues

```bash
# Test database connection
mysql -u woodpecker -p -h localhost woodpecker
# or
psql -U woodpecker -h localhost woodpecker
```

## Upgrading

To upgrade to a new version:

```bash
# Stop services
sudo systemctl stop woodpecker-agent
sudo systemctl stop woodpecker-server

# Backup database
sudo cp /var/lib/woodpecker/woodpecker.sqlite /var/lib/woodpecker/woodpecker.sqlite.backup

# Download and install new packages (see Step 1-2)

# Start services
sudo systemctl start woodpecker-server
sudo systemctl start woodpecker-agent
```

## Security Considerations

1. **File Permissions**: Ensure environment files are readable only by root and the woodpecker user
2. **Database**: Use strong passwords and restrict network access
3. **Secrets**: Store `WOODPECKER_AGENT_SECRET` and OAuth credentials securely
4. **HTTPS**: Always use HTTPS in production with valid certificates
5. **Firewall**: Restrict access to port 9000 (gRPC) to only agent IPs
6. **Updates**: Keep the system and Woodpecker updated regularly

## Community Packages

:::info
These packages are not maintained by the Woodpecker developers. Please reach out to the respective maintainers for support.
:::

- [Alpine (Edge)](https://pkgs.alpinelinux.org/packages?name=woodpecker&branch=edge)
- [Arch Linux](https://archlinux.org/packages/?q=woodpecker)
- [openSUSE](https://software.opensuse.org/package/woodpecker)
- [YunoHost](https://apps.yunohost.org/app/woodpecker)
- [Cloudron](https://www.cloudron.io/store/org.woodpecker_ci.cloudronapp.html)
- [Easypanel](https://easypanel.io/docs/templates/woodpeckerci)

## Next Steps

- Configure your first pipeline: [Pipeline Documentation](/docs/usage/pipeline-syntax)
- Learn about secrets management: [Secrets Guide](/docs/usage/secrets)
- Set up multiple agents for scaling: [Agent Configuration](/docs/administration/agent-config)
