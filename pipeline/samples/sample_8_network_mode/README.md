# Example

Compile the yaml to the intermediate representation:

```sh
pipec compile
```

Execute the intermediate representation:

```sh
pipec exec
```

This example shows how to use the network_mode option to use the network defined by other container.
This is useful for example to allow the CI to connect with servers behind a VPN.

Before to start you need to create a container that connects to the VPN (using one of the openvpn client images like <https://github.com/ekristen/docker-openvpn-client>).
