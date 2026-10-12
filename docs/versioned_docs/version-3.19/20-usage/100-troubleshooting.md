# Troubleshooting

## How to debug clone issues

(And what to do with an error message like `fatal: could not read Username for 'https://<url>': No such device or address`)

This error can have multiple causes. If you use internal repositories you might have to enable `WOODPECKER_AUTHENTICATE_PUBLIC_REPOS`:

```ini
WOODPECKER_AUTHENTICATE_PUBLIC_REPOS=true
```

If that does not work, try to make sure the container can reach your git server. In order to do that disable git checkout and make the container "hang":

```yaml
skip_clone: true

steps:
  build:
    image: debian:stable-backports
    commands:
      - apt update
      - apt install -y inetutils-ping wget
      - ping -c 4 git.example.com
      - wget git.example.com
      - sleep 9999999
```

Get the container id using `docker ps` and copy the id from the first column. Enter the container with: `docker exec -it 1234asdf  bash` (replace `1234asdf` with the docker id). Then try to clone the git repository with the commands from the failing pipeline:

```bash
git init
git remote add origin https://git.example.com/username/repo.git
git fetch --no-tags origin +refs/heads/branch:
```

(replace the url AND the branch with the correct values, use your username and password as log in values)

## SELinux Issues

When running Woodpecker on systems with SELinux enabled (such as RHEL, CentOS, Fedora, or other Enterprise Linux distributions), SELinux may prevent the agent from accessing the Docker socket.

### Symptoms

If SELinux is blocking access, you may see errors like:

```text
permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock
```

### Solutions

There are several ways to resolve this:

#### Option 1: Set SELinux to Permissive Mode (For Testing Only)

Set SELinux to permissive mode temporarily to verify it's the issue:

```bash
setenforce 0
```

To permanently set SELinux to permissive mode:

```bash
# Edit /etc/selinux/config
SELINUX=permissive
```

#### Option 2: Configure SELinux Policy (Recommended)

Create a custom SELinux policy to allow Woodpecker agent to access Docker:

```bash
# Generate the policy module
ausearch -c 'docker' -avc | audit2allow -R -o woodpecker-docker.te
# Build the policy module
checkmodule -M -m -o woodpecker-docker.mod woodpecker-docker.te
semodule_package -o woodpecker-docker.pp -m woodpecker-docker.mod
# Load the policy module
semodule -i woodpecker-docker.pp
```

#### Option 3: Use Docker Volume with SELinux Options

When using Docker Compose or Docker, add the `:z` or `:Z` option to volume mounts:

```yaml
volumes:
  - /var/run/docker.sock:/var/run/docker.sock:z
```

The `:z` option tells Docker to automatically relabel the volume content for SELinux. Use `:Z` with caution as it relabels the volume exclusively for this container.

#### Option 4: Use Podman (Alternative)

If you prefer to avoid SELinux configuration issues, consider using Podman instead of Docker, as it has better SELinux integration.
