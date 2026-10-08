# Documentation

The documentation is using docusaurus as framework. You can learn more about it from its [official documentation](https://docusaurus.io/docs/).

If you only want to change some text it probably is enough if you just search for the corresponding [Markdown](https://www.markdownguide.org/basic-syntax/) file inside the `docs/docs/` folder and adjust it. If you want to change larger parts and test the rendered documentation you can run docusaurus locally. Similarly to the UI you need to install [Node.js and pnpm](./01-getting-started.md#install-nodejs--pnpm). After that you can run and build docusaurus locally by using the following commands:

```bash
cd docs/

pnpm install

# build plugins used by the docs
pnpm build:woodpecker-plugins

# start docs with hot-reloading, so you can change the docs and directly see the changes in the browser without reloading it manually
pnpm start

# or build the docs to deploy it to some static page hosting
pnpm build
```

## Versions

The folder `docs/docs/` contains the documentation of the upcoming version (`next`). This is where changes normally go. The documentation of each released version is a snapshot in `docs/versioned_docs/`. Only change it in addition if a fix should also show up for an already released version.

### Creating a new version

The snapshot of a new version is created by a release manager in a dedicated pull request, right before the new version of Woodpecker gets released:

```bash
# generate the CLI docs first
make generate-docs

# switch to docs and install dependencys
cd docs/; pnpm i

# generate version snapshot and format
pnpm docusaurus docs:version x.x; pnpm format
```

:::warning
Always run `make generate-docs` before you create the snapshot. The CLI docs (`docs/docs/40-cli.md`) are not part of the repository,
they are generated from the code each time the docs of `next` get built. The snapshot only contains the files that exist at that moment.
So without this step the CLI docs are missing in the new version or, even worse, are outdated.
:::

## Formatting and linting

The CI checks the documentation, so run these checks before you open a pull request:

- [Prettier](https://prettier.io/) formats the files. It has to be run inside of `docs/`, as the Prettier config in the root of the repository ignores this folder:

  ```bash
  cd docs/

  # check the formatting
  pnpm format:check

  # fix the formatting
  pnpm format
  ```

- [markdownlint](https://github.com/DavidAnson/markdownlint) checks the Markdown syntax. It is part of the [`pre-commit`](./01-getting-started.md#install-pre-commit-optional) hooks.
- [CSpell](https://cspell.org/) checks the spelling and can be executed with `make spellcheck`. Unknown but correct words can be added to `.cspell.json` or allowed for a single file with a comment like `<!-- cspell:ignore someword -->`.
