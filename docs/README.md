# Website

This website is built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.

### Installation

```
pnpm install
```

### Local Development

```
pnpm start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

### Build

```
pnpm build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

### Deployment

Deployment happen via [CI](https://github.com/woodpecker-ci/woodpecker/blob/d59fdb4602bfdd0d00078716ba61b05c02cbd1af/.woodpecker/docs.yml#L8-L30) to [woodpecker-ci.org](https://woodpecker-ci.org).

To manually build the website and push it exec:

```sh
GIT_USER=woodpecker-bot USE_SSH=true DEPLOYMENT_BRANCH=master pnpm deploy
```
