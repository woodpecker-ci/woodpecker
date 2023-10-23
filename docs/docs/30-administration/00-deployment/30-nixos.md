# NixOS

:::info
Note that this module is not maintained by the woodpecker-developers.
If you experience issues please open a bug report in the [nixpkgs repo](https://github.com/NixOS/nixpkgs/issues/new/choose) where the module is maintained.
:::

The NixOS install is in theory quite similar to the binary install and supports multiple backends.
In practice, the settings are specified declaratively in the NixOS configuration and no manual steps need to be taken.

## General Configuration

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

All configuration options can be found via [NixOS Search](https://search.nixos.org/options?channel=unstable&size=200&sort=relevance&query=woodpecker)

## Tips and tricks

There are some resources on how to utilize Woodpecker more effectively with NixOS on the [Awesome Woodpecker](../../92-awesome.md) page, like using the runners nix-store in the pipeline
