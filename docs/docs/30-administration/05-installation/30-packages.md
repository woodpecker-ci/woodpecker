# Binary Installation Guide

This guide walks you through installing Woodpecker CI from binary packages on Linux systems. Binary installation gives you full control over the deployment and is ideal for bare metal servers, VMs, or when you prefer not to use Docker.

## Prerequisites

Before starting the installation, ensure your system meets the following requirements:

### System Requirements

- **Operating System**: Linux (Ubuntu 20.04+, Debian 11+, CentOS 8+, RHEL 8+, or compatible)
- **Architecture**: x86_64 (amd64) or aarch64 (arm64)
- **Memory**: Minimum 1 GB RAM (2 GB recommended for production)
- **Disk Space**: At least 10 GB free space (more for build artifacts)
- **Network**: Outbound HTTPS access (port 443) to your Git forge

### Required Software

- `curl` or `wget` for downloading packages
- `sudo` privileges for installation
- A supported database (SQLite, MySQL, or PostgreSQL)

### Optional but Recommended

- A domain name and SSL certificate for HTTPS
- A reverse proxy (Nginx or Apache) for production deployments

## Step-by-step Installation

### Step 1: Create the Woodpecker User

Create a dedicated system user for running Woodpecker services:

```shell
sudo useradd --system --user-group --create-home --home-dir /var/lib/woodpecker woodpecker
```

This creates a system user with:
- A home directory at `/var/lib/woodpecker` for storing data
- No shell access for security
- A dedicated group for file permissions

### Step 2: Download and Install Packages

Download the latest Woodpecker packages from GitHub releases:

```shell
# Get the latest release version
RELEASE_VERSION=$(curl -s https://api.github.com/repos/woodpecker-ci/woodpecker/releases/latest | grep -Po '"tag_name":\s"v\K[^"]+')

# For Debian/Ubuntu (x86_64)
curl -fLO "https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-server_${RELEASE_VERSION}_amd64.deb"
curl -fLO "https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-agent_${RELEASE_VERSION}_amd64.deb"
curl -fLO "https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-cli_${RELEASE_VERSION}_amd64.deb"
sudo apt --fix-broken install ./woodpecker-{server,agent,cli}_${RELEASE_VERSION}_amd64.deb

# For CentOS/RHEL/Rocky Linux (x86_64)
sudo dnf install https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-server-${RELEASE_VERSION}.x86_64.rpm
sudo dnf install https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-agent-${RELEASE_VERSION}.x86_64.rpm
sudo dnf install https://github.com/woodpecker-ci/woodpecker/releases/download/v${RELEASE_VERSION}/woodpecker-cli-${RELEASE_VERSION}.x86_64.rpm

# For aarch64 (ARM64), replace 'amd64' or 'x86_64' with 'arm64' or 'aarch64' respectively
```

### Step 3: Set Up the Database

Woodpecker supports SQLite, MySQL, and PostgreSQL. Choose the one that fits your needs:

#### Option A: SQLite (Simplest - Good for Small Setups)

SQLite is suitable for small teams or testing. The database will be created automatically on first run.

```shell
# Create the database directory
sudo mkdir -p /var/lib/woodpecker
sudo chown -R woodpecker:woodpecker /var/lib/woodpecker
```

No additional setup required! Woodpecker will create the database file automatically.

#### Option B: MySQL/MariaDB (Recommended for Production)

```shell
# Install MySQL/MariaDB server (if not already installed)
# Ubuntu/Debian:
sudo apt install mysql-server

# CentOS/RHEL:
sudo dnf install mysql-server

# Create the database and user
sudo mysql -u root -p
```

Run these SQL commands:

```sql
CREATE DATABASE woodpecker CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'woodpecker'@'localhost' IDENTIFIED BY 'your_secure_password_here';
GRANT ALL PRIVILEGES ON woodpecker.* TO 'woodpecker'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

#### Option C: PostgreSQL (Alternative for Production)

```shell
# Install PostgreSQL (if not already installed)
# Ubuntu/Debian:
sudo apt install postgresql postgresql-contrib

# CentOS/RHEL:
sudo dnf install postgresql-server postgresql-contrib

# Initialize and start PostgreSQL
sudo postgresql-setup --initdb
sudo systemctl enable --now postgresql

# Create the database and user
sudo -u postgres psql
```

Run these SQL commands:

```sql
CREATE DATABASE woodpecker;
CREATE USER woodpecker WITH ENCRYPTED PASSWORD 'your_secure_password_here';
GRANT ALL PRIVILEGES ON DATABASE woodpecker TO woodpecker;
\q
```

### Step 4: Configure the Server

Copy the example environment file and configure it:

```shell
sudo cp /etc/woodpecker/woodpecker-server.env.example /etc/woodpecker/woodpecker-server.env
sudo chmod 600 /etc/woodpecker/woodpecker-server.env
```

Edit `/etc/woodpecker/woodpecker-server.env` with your settings:

```ini
# Server configuration
WOODPECKER_HOST=https://ci.yourdomain.com
WOODPECKER_SERVER_ADDR=:8000

# Database configuration (choose one)

# For SQLite (default):
WOODPECKER_DATABASE_DRIVER=sqlite3
WOODPECKER_DATABASE_DATASOURCE=/var/lib/woodpecker/woodpecker.sqlite

# For MySQL/MariaDB:
# WOODPECKER_DATABASE_DRIVER=mysql
# WOODPECKER_DATABASE_DATASOURCE=woodpecker:your_secure_password_here@tcp(localhost:3306)/woodpecker?parseTime=true

# For PostgreSQL:
# WOODPECKER_DATABASE_DRIVER=postgres
# WOODPECKER_DATABASE_DATASOURCE=postgres://woodpecker:your_secure_password_here@localhost:5432/woodpecker?sslmode=disable

# Security - Generate a strong random secret
WOODPECKER_AGENT_SECRET=$(openssl rand -hex 32)

# Forge configuration (GitHub example)
WOODPECKER_GITHUB=true
WOODPECKER_GITHUB_CLIENT=your_github_client_id
WOODPECKER_GITHUB_SECRET=your_github_client_secret

# Admin users (comma-separated list of usernames)
WOODPECKER_ADMIN_USER=your_github_username

# TLS/HTTPS (disable if using reverse proxy)
# WOODPECKER_LETS_ENCRYPT=true
# WOODPECKER_LETS_ENCRYPT_EMAIL=admin@yourdomain.com
```

### Step 5: Configure systemd Services

The packages install basic systemd service files. For production, use these hardened configurations:

#### Hardened Server Service

Create `/etc/systemd/system/woodpecker-server.service`:

```ini
[Unit]
Description=Woodpecker CI Server
Documentation=https://woodpecker-ci.org/docs/administration/server-config
Requires=network.target
After=network-online.target
Wants=network-online.target
ConditionFileNotEmpty=/etc/woodpecker/woodpecker-server.env

[Service]
Type=simple
User=woodpecker
Group=woodpecker
WorkingDirectory=/var/lib/woodpecker

# Environment
EnvironmentFile=/etc/woodpecker/woodpecker-server.env
Environment="GIN_MODE=release"

# Execution
ExecStart=/usr/local/bin/woodpecker-server
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
RestartSec=5

# Resource limits
LimitNOFILE=65536
MemoryMax=2G

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/woodpecker
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictSUIDSGID=true
RestrictRealtime=true
RestrictNamespaces=true
LockPersonality=true
MemoryDenyWriteExecute=true
SystemCallArchitectures=native

[Install]
WantedBy=multi-user.target
```

#### Hardened Agent Service

Create `/etc/systemd/system/woodpecker-agent.service`:

```ini
[Unit]
Description=Woodpecker CI Agent
Documentation=https://woodpecker-ci.org/docs/administration/agent-config
Requires=network.target
After=network-online.target woodpecker-server.service
Wants=network-online.target
ConditionFileNotEmpty=/etc/woodpecker/woodpecker-agent.env

[Service]
Type=simple
User=woodpecker
Group=woodpecker
WorkingDirectory=/var/lib/woodpecker

# Environment
EnvironmentFile=/etc/woodpecker/woodpecker-agent.env

# Execution
ExecStart=/usr/local/bin/woodpecker-agent
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
RestartSec=5

# Resource limits
LimitNOFILE=65536
MemoryMax=1G

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/woodpecker
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictSUIDSGID=true
RestrictRealtime=true
RestrictNamespaces=true
LockPersonality=true
SystemCallArchitectures=native

[Install]
WantedBy=multi-user.target
```

Reload systemd and enable services:

```shell
sudo systemctl daemon-reload
sudo systemctl enable woodpecker-server woodpecker-agent
```

### Step 6: Configure the Agent

Copy and configure the agent environment file:

```shell
sudo cp /etc/woodpecker/woodpecker-agent.env.example /etc/woodpecker/woodpecker-agent.env
sudo chmod 600 /etc/woodpecker/woodpecker-agent.env
```

Edit `/etc/woodpecker/woodpecker-agent.env`:

```ini
# Server connection
WOODPECKER_SERVER=localhost:9000

# Must match WOODPECKER_AGENT_SECRET on the server
WOODPECKER_AGENT_SECRET=your_agent_secret_here

# Optional: limit concurrent workflows
# WOODPECKER_MAX_WORKFLOWS=2

# Optional: set custom hostname for this agent
# WOODPECKER_HOSTNAME=agent-01
```

### Step 7: Start the Services

Start and verify the services:

```shell
# Start the server
sudo systemctl start woodpecker-server

# Check server status
sudo systemctl status woodpecker-server
sudo journalctl -u woodpecker-server -f

# Once the server is running, start the agent
sudo systemctl start woodpecker-agent

# Check agent status
sudo systemctl status woodpecker-agent
sudo journalctl -u woodpecker-agent -f
```

Verify the installation by accessing `http://your-server-ip:8000` (or your configured domain).

## Reverse Proxy Configuration

For production deployments, it's recommended to use a reverse proxy with HTTPS.

### Nginx Configuration

Install Nginx:

```shell
# Ubuntu/Debian
sudo apt install nginx

# CentOS/RHEL
sudo dnf install nginx
```

Create `/etc/nginx/sites-available/woodpecker` (Ubuntu/Debian) or `/etc/nginx/conf.d/woodpecker.conf` (CentOS/RHEL):

```nginx
upstream woodpecker {
    server 127.0.0.1:8000;
}

server {
    listen 80;
    server_name ci.yourdomain.com;

    # Redirect HTTP to HTTPS
    location / {
        return 301 https://$server_name$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name ci.yourdomain.com;

    # SSL configuration (adjust paths as needed)
    ssl_certificate /etc/letsencrypt/live/ci.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/ci.yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Logging
    access_log /var/log/nginx/woodpecker-access.log;
    error_log /var/log/nginx/woodpecker-error.log;

    # Proxy settings
    location / {
        proxy_pass http://woodpecker;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
    }

    # WebSocket support for real-time updates
    location /ws {
        proxy_pass http://woodpecker;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 86400;
    }
}
```

Enable the site and reload Nginx:

```shell
# Ubuntu/Debian
sudo ln -s /etc/nginx/sites-available/woodpecker /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# CentOS/RHEL
sudo nginx -t
sudo systemctl reload nginx
```

### Apache Configuration

Install Apache:

```shell
# Ubuntu/Debian
sudo apt install apache2

# CentOS/RHEL
sudo dnf install httpd
```

Enable required modules:

```shell
# Ubuntu/Debian
sudo a2enmod proxy proxy_http proxy_wstunnel ssl rewrite headers

# CentOS/RHEL - modules are typically enabled by default
```

Create `/etc/apache2/sites-available/woodpecker.conf` (Ubuntu/Debian) or `/etc/httpd/conf.d/woodpecker.conf` (CentOS/RHEL):

```apache
<VirtualHost *:80>
    ServerName ci.yourdomain.com
    Redirect permanent / https://ci.yourdomain.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName ci.yourdomain.com
    ServerAdmin admin@yourdomain.com

    # SSL Configuration
    SSLEngine on
    SSLCertificateFile /etc/letsencrypt/live/ci.yourdomain.com/cert.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/ci.yourdomain.com/privkey.pem
    SSLCertificateChainFile /etc/letsencrypt/live/ci.yourdomain.com/chain.pem

    # Security headers
    Header always set X-Frame-Options "SAMEORIGIN"
    Header always set X-Content-Type-Options "nosniff"
    Header always set X-XSS-Protection "1; mode=block"
    Header always set Referrer-Policy "strict-origin-when-cross-origin"

    # Proxy configuration
    ProxyPreserveHost On
    ProxyRequests Off

    # WebSocket support
    RewriteEngine On
    RewriteCond %{HTTP:Upgrade} websocket [NC]
    RewriteCond %{HTTP:Connection} upgrade [NC]
    RewriteRule ^/?(.*) "ws://127.0.0.1:8000/$1" [P,L]

    # Regular proxy
    ProxyPass / http://127.0.0.1:8000/
    ProxyPassReverse / http://127.0.0.1:8000/

    # Logging
    ErrorLog ${APACHE_LOG_DIR}/woodpecker-error.log
    CustomLog ${APACHE_LOG_DIR}/woodpecker-access.log combined
</VirtualHost>
```

Enable the site and reload Apache:

```shell
# Ubuntu/Debian
sudo a2ensite woodpecker
sudo apache2ctl configtest
sudo systemctl reload apache2

# CentOS/RHEL
sudo httpd -t
sudo systemctl reload httpd
```

## Troubleshooting

### Server won't start

**Problem:** `systemctl start woodpecker-server` fails

**Solutions:**

1. Check the logs:
   ```shell
   sudo journalctl -u woodpecker-server -n 50
   ```

2. Verify the environment file exists and is readable:
   ```shell
   sudo ls -la /etc/woodpecker/woodpecker-server.env
   sudo test -r /etc/woodpecker/woodpecker-server.env && echo "Readable" || echo "Not readable"
   ```

3. Check database connectivity:
   ```shell
   # For MySQL
   mysql -u woodpecker -p -e "SELECT 1;" woodpecker

   # For PostgreSQL
   sudo -u woodpecker psql -h localhost -d woodpecker -c "SELECT 1;"
   ```

4. Ensure the working directory exists and is writable:
   ```shell
   sudo ls -la /var/lib/woodpecker
   sudo chown -R woodpecker:woodpecker /var/lib/woodpecker
   ```

### Agent can't connect to server

**Problem:** Agent shows connection errors

**Solutions:**

1. Verify the server is running and listening:
   ```shell
   sudo ss -tlnp | grep woodpecker
   sudo systemctl status woodpecker-server
   ```

2. Check the agent secret matches the server's secret:
   ```shell
   sudo grep WOODPECKER_AGENT_SECRET /etc/woodpecker/woodpecker-server.env
   sudo grep WOODPECKER_AGENT_SECRET /etc/woodpecker/woodpecker-agent.env
   ```

3. Verify network connectivity:
   ```shell
   telnet localhost 9000
   ```

### GitHub authentication issues

**Problem:** Can't log in with GitHub

**Solutions:**

1. Verify OAuth application settings in GitHub match your WOODPECKER_HOST
2. Check that the GitHub Client ID and Secret are correct:
   ```shell
   sudo grep "WOODPECKER_GITHUB" /etc/woodpecker/woodpecker-server.env
   ```
3. Ensure the callback URL in GitHub OAuth settings is: `https://ci.yourdomain.com/authorize`

### Database connection errors

**Problem:** "unable to open database file" or connection refused errors

**Solutions:**

1. **SQLite:** Ensure the directory exists and is writable:
   ```shell
   sudo mkdir -p /var/lib/woodpecker
   sudo chown -R woodpecker:woodpecker /var/lib/woodpecker
   sudo chmod 750 /var/lib/woodpecker
   ```

2. **MySQL/MariaDB:** Check that the database and user exist:
   ```shell
   sudo mysql -e "SHOW DATABASES; SELECT user FROM mysql.user;"
   ```

3. **PostgreSQL:** Verify authentication method in `pg_hba.conf`:
   ```shell
   sudo grep woodpecker /etc/postgresql/*/main/pg_hba.conf
   ```

## Security Considerations

### File Permissions

Ensure sensitive files have proper permissions:

```shell
# Environment files should be readable only by root and woodpecker user
sudo chmod 600 /etc/woodpecker/woodpecker-*.env
sudo chown root:woodpecker /etc/woodpecker/woodpecker-*.env

# Database directory should be owned by woodpecker user
sudo chown -R woodpecker:woodpecker /var/lib/woodpecker
sudo chmod 750 /var/lib/woodpecker
```

### Secrets Management

1. **Generate strong secrets:**
   ```shell
   openssl rand -hex 32
   ```

2. **Use a secrets management tool** (optional but recommended):
   - HashiCorp Vault
   - AWS Secrets Manager
   - Azure Key Vault
   - 1Password Secrets Automation

3. **Rotate secrets regularly**, especially after personnel changes

### Network Security

1. **Use HTTPS in production** - never expose Woodpecker over plain HTTP
2. **Place behind a firewall** - restrict access to necessary ports only
3. **Use a VPN or private network** for agent-to-server communication if possible
4. **Enable fail2ban** to prevent brute force attacks:
   ```shell
   sudo apt install fail2ban  # or dnf install fail2ban
   ```

### Regular Updates

Subscribe to security advisories:
- Watch the [woodpecker-ci/woodpecker](https://github.com/woodpecker-ci/woodpecker) repository
- Join the community Discord/Matrix for announcements
- Check the [security policy](https://github.com/woodpecker-ci/woodpecker/security)

## Upgrade Instructions

### Before Upgrading

1. **Backup your database:**

   **SQLite:**
   ```shell
   sudo systemctl stop woodpecker-server woodpecker-agent
   sudo cp /var/lib/woodpecker/woodpecker.sqlite /var/lib/woodpecker/woodpecker.sqlite.backup.$(date +%Y%m%d)
   ```

   **MySQL/MariaDB:**
   ```shell
   mysqldump -u woodpecker -p woodpecker > woodpecker-backup-$(date +%Y%m%d).sql
   ```

   **PostgreSQL:**
   ```shell
   pg_dump -U woodpecker -h localhost woodpecker > woodpecker-backup-$(date +%Y%m%d).sql
   ```

2. **Backup configuration files:**
   ```shell
   sudo cp -r /etc/woodpecker /etc/woodpecker.backup.$(date +%Y%m%d)
   ```

3. **Review release notes** for breaking changes

### Performing the Upgrade

1. **Stop the services:**
   ```shell
   sudo systemctl stop woodpecker-server woodpecker-agent
   ```

2. **Download and install new packages** (same commands as initial installation)

3. **Review and update configuration** if needed (check release notes)

4. **Start the services:**
   ```shell
   sudo systemctl start woodpecker-server
   # Wait for the server to be ready
   sleep 5
   sudo systemctl start woodpecker-agent
   ```

5. **Verify the upgrade:**
   ```shell
   sudo systemctl status woodpecker-server woodpecker-agent
   sudo journalctl -u woodpecker-server -n 20
   ```

### Rollback Procedure

If something goes wrong:

1. **Stop the services:**
   ```shell
   sudo systemctl stop woodpecker-server woodpecker-agent
   ```

2. **Restore the database from backup**

3. **Reinstall the previous version packages** (download from GitHub releases)

4. **Restore configuration files**

5. **Start the services:**
   ```shell
   sudo systemctl start woodpecker-server woodpecker-agent
   ```

## Community Packages

:::info
Woodpecker itself is not responsible for creating these packages. Please reach out to the people responsible for packaging Woodpecker for the individual distributions.
:::

- [Alpine (Edge)](https://pkgs.alpinelinux.org/packages?name=woodpecker&branch=edge&repo=&arch=&maintainer=)
- [Arch Linux](https://archlinux.org/packages/?q=woodpecker)
- [openSUSE](https://software.opensuse.org/package/woodpecker)
- [YunoHost](https://apps.yunohost.org/app/woodpecker)
- [Cloudron](https://www.cloudron.io/store/org.woodpecker_ci.cloudronapp.html)
- [Easypanel](https://easypanel.io/docs/templates/woodpeckerci)
- [Homebrew](https://formulae.brew.sh/formula/woodpecker-cli) (CLI only)

### NixOS

:::info
This module is not maintained by the Woodpecker developers.
If you experience issues please open a bug report in the [nixpkgs repo](https://github.com/NixOS/nixpkgs/issues/new/choose) where the module is maintained.
:::

In theory, the NixOS installation is very similar to the binary installation and supports multiple backends.
In practice, the settings are specified declaratively in the NixOS configuration and no manual steps need to be taken.

<!-- cspell:words Optimisation -->

```nix
{ config
, ...
}:
let
  domain = "woodpecker.example.org";
in
{
  # This automatically sets up certificates via let's encrypt
  security.acme.defaults.email = "acme@example.com";
  security.acme.acceptTerms = true;

  # Setting up a nginx proxy that handles tls for us
  services.nginx = {
    enable = true;
    openFirewall = true;
    recommendedTlsSettings = true;
    recommendedOptimisation = true;
    recommendedProxySettings = true;
    virtualHosts."${domain}" = {
      enableACME = true;
      forceSSL = true;
      locations."/".proxyPass = "http://localhost:3007";
    };
  };

  services.woodpecker-server = {
    enable = true;
    environment = {
      WOODPECKER_HOST = "https://${domain}";
      WOODPECKER_SERVER_ADDR = ":3007";
      WOODPECKER_OPEN = "true";
    };
    # You can pass a file with env vars to the system it could look like:
    # WOODPECKER_AGENT_SECRET=XXXXXXXXXXXXXXXXXXXXXX
    environmentFile = "/path/to/my/secrets/file";
  };

  # This sets up a woodpecker agent
  services.woodpecker-agents.agents."docker" = {
    enable = true;
    # We need this to talk to the podman socket
    extraGroups = [ "podman" ];
    environment = {
      WOODPECKER_SERVER = "localhost:9000";
      WOODPECKER_MAX_WORKFLOWS = "4";
      DOCKER_HOST = "unix:///run/podman/podman.sock";
      WOODPECKER_BACKEND = "docker";
    };
    # Same as with woodpecker-server
    environmentFile = [ "/var/lib/secrets/woodpecker.env" ];
  };

  # Here we setup podman and enable dns
  virtualisation.podman = {
    enable = true;
    defaultNetwork.settings = {
      dns_enabled = true;
    };
  };
  # This is needed for podman to be able to talk over dns
  networking.firewall.interfaces."podman0" = {
    allowedUDPPorts = [ 53 ];
    allowedTCPPorts = [ 53 ];
  };
}
```

All configuration options can be found via [NixOS Search](https://search.nixos.org/options?channel=unstable&size=200&sort=relevance&query=woodpecker). There are also some additional resources on how to utilize Woodpecker more effectively with NixOS on the [Awesome Woodpecker](/awesome) page, like using the runners nix-store in the pipeline.
