# GitHub

Woodpecker comes with built-in support for GitHub and GitHub Enterprise.
To use Woodpecker with GitHub the following environment variables should be set for the server component:

```ini title=".env"
WOODPECKER_GITHUB=true
WOODPECKER_GITHUB_CLIENT=YOUR_GITHUB_CLIENT_ID
WOODPECKER_GITHUB_SECRET=YOUR_GITHUB_CLIENT_SECRET
```

You will get these values from GitHub when you register your OAuth application.
To do so, go to Settings -> Developer Settings -> GitHub Apps -> New Oauth2 App.

:::warning
Do not use a "GitHub App" instead of an Oauth2 app as the former will not work correctly with Woodpecker right now (because user access tokens are not being refreshed automatically)
:::

## App Settings

- Name: An arbitrary name for your App
- Homepage URL: The URL of your Woodpecker instance
- Callback URL: `https://<your-woodpecker-instance>/authorize`
- (optional) Upload the Woodpecker Logo: <https://avatars.githubusercontent.com/u/84780935?s=200&v=4>

## Client Secret Creation

After your App has been created, you can generate a client secret.
Use this one for the `WOODPECKER_GITHUB_SECRET` environment variable.

## Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

### `WOODPECKER_GITHUB`

> Default: `false`

Enables the GitHub driver.

### `WOODPECKER_GITHUB_URL`

> Default: `https://github.com`

Configures the GitHub server address.

### `WOODPECKER_GITHUB_CLIENT`

> Default: empty

Configures the GitHub OAuth client id to authorize access.

### `WOODPECKER_GITHUB_CLIENT_FILE`

> Default: empty

Read the value for `WOODPECKER_GITHUB_CLIENT` from the specified filepath.

### `WOODPECKER_GITHUB_SECRET`

> Default: empty

Configures the GitHub OAuth client secret. This is used to authorize access.

### `WOODPECKER_GITHUB_SECRET_FILE`

> Default: empty

Read the value for `WOODPECKER_GITHUB_SECRET` from the specified filepath.

### `WOODPECKER_GITHUB_MERGE_REF`

> Default: `true`

### `WOODPECKER_GITHUB_SKIP_VERIFY`

> Default: `false`

Configure if SSL verification should be skipped.
