# Distribution packages

## Official packages

- DEB
- RPM

The pre-built packages are available on the [GitHub releases](https://github.com/woodpecker-ci/woodpecker/releases/latest) page. The packages can be installed using the package manager of your distribution.

```Shell
RELEASE_VERSION=$(curl -s https://api.github.com/repos/woodpecker-ci/woodpecker/releases/latest | grep -Po '"tag_name":\s"v\K[^"]+')

# Debian/Ubuntu (x86_64)
curl -fLOOO "https://github.com/woodpecker-ci/woodpecker/releases/download/${RELEASE_VERSION}/woodpecker-{server,agent,cli}_${RELEASE_VERSION#v}_amd64.deb"
sudo apt --fix-broken install ./woodpecker-{server,agent,cli}_${RELEASE_VERSION#v}_amd64.deb

# CentOS/RHEL (x86_64)
sudo dnf install https://github.com/woodpecker-ci/woodpecker/releases/download/${RELEASE_VERSION}/woodpecker-{server,agent,cli}-${RELEASE_VERSION#v}.x86_64.rpm
```

The package installation will create a systemd service file for the Woodpecker server and agent along with an example environment file. To configure the server, copy the example environment file `/etc/woodpecker/woodpecker-server.env.example` to `/etc/woodpecker/woodpecker-server.env` and adjust the values.

```ini title="/usr/local/lib/systemd/system/woodpecker-server.service"
[Unit]
Description=WoodpeckerCI server
Documentation=https://woodpecker-ci.org/docs/administration/server-config
Requires=network.target
After=network.target
ConditionFileNotEmpty=/etc/woodpecker/woodpecker-server.env
ConditionPathExists=/etc/woodpecker/woodpecker-server.env

[Service]
Type=simple
EnvironmentFile=/etc/woodpecker/woodpecker-server.env
User=woodpecker
Group=woodpecker
ExecStart=/usr/local/bin/woodpecker-server
WorkingDirectory=/var/lib/woodpecker/
StateDirectory=woodpecker

[Install]
WantedBy=multi-user.target
```

```shell title="/etc/woodpecker/woodpecker-server.env"
WOODPECKER_OPEN=true
WOODPECKER_HOST=${WOODPECKER_HOST}
WOODPECKER_GITHUB=true
WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
WOODPECKER_AGENT_SECRET=${WOODPECKER_AGENT_SECRET}
```

After installing the agent, copy the example environment file `/etc/woodpecker/woodpecker-agent.env.example` to `/etc/woodpecker/woodpecker-agent.env` and adjust the values as well. The agent will automatically register itself with the server.

```ini title="/usr/local/lib/systemd/system/woodpecker-agent.service"
[Unit]
Description=WoodpeckerCI agent
Documentation=https://woodpecker-ci.org/docs/administration/agent-config
Requires=network.target
After=network.target
ConditionFileNotEmpty=/etc/woodpecker/woodpecker-agent.env
ConditionPathExists=/etc/woodpecker/woodpecker-agent.env

[Service]
Type=simple
EnvironmentFile=/etc/woodpecker/woodpecker-agent.env
User=woodpecker
Group=woodpecker
ExecStart=/usr/local/bin/woodpecker-agent
WorkingDirectory=/var/lib/woodpecker/
StateDirectory=woodpecker

[Install]
WantedBy=multi-user.target
```

```shell title="/etc/woodpecker/woodpecker-agent.env"
WOODPECKER_SERVER=localhost:9000
WOODPECKER_AGENT_SECRET=${WOODPECKER_AGENT_SECRET}
```

## Community packages

:::info
Woodpecker itself is not responsible for creating these packages. Please reach out to the people responsible for packaging Woodpecker for the individual distributions.
:::

- [Alpine (Edge)](https://pkgs.alpinelinux.org/packages?name=woodpecker&branch=edge&repo=&arch=&maintainer=)
- [Arch Linux](https://archlinux.org/packages/?q=woodpecker)
- [openSUSE](https://software.opensuse.org/package/woodpecker)
- [YunoHost](https://apps.yunohost.org/app/woodpecker)
- [Cloudron](https://www.cloudron.io/store/org.woodpecker_ci.cloudronapp.html)

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
  security.acme.certs."${domain}" = { };

  # Setting up a nginx proxy that handles tls for us
  networking.firewall.allowedTCPPorts = [ 80 443 ];
  services.nginx = {
    enable = true;
    recommendedTlsSettings = true;
    recommendedOptimisation = true;
    recommendedProxySettings = true;
    virtualHosts."${domain}" = {
      enableACME = true;
      forceSSL = true;
      locations."/" = {
        proxyPass = "http://localhost:3007";
      };
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
