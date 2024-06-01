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
