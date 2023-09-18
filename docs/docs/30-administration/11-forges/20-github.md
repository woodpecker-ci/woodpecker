# GitHub

Woodpecker comes with built-in support for GitHub and GitHub Enterprise.
To use Woodpecker with GitHub the following environment variables should be set for the server component:

```sh
WOODPECKER_GITHUB=true
WOODPECKER_GITHUB_CLIENT=YOUR_GITHUB_CLIENT_ID
WOODPECKER_GITHUB_SECRET=YOUR_GITHUB_CLIENT_SECRET
```

You will get these values from GitHub when you register your application.
To do so, go to Settings -> Developer Settings -> GitHub Apps -> New GitHub App.

## App Settings

- Name: An arbitrary name for your App
- Homepage URL: The URL of your Woodpecker instance
- Callback URL: `https://<your-woodpecker-instance>/authorize`
- Leave "Request user authorization (OAuth) during installation" and "Enable Device Flow" unchecked
- Leave "Webhook" and "Post Installation" fields empty
- (optional) Upload the Woodpecker Logo: https://avatars.githubusercontent.com/u/84780935?s=200&v=4

## App Permissions

The app must be granted the following permissions (under App Settings -> Permissions):

Repository:

- Commit statuses: Read & write
- Contents: Read & write
- Deployments: Read & write
- Metadata: Read-only
- Pull requests: Read & write
- Secrets: Read & write
- Webhooks: Read & write

Organization:

- Members: Read-only

Account:

- Email addresses: Read-only

## Client Secret Creation

After your App has been created, you can generate a client secret.
Use this one for the `WOODPECKER_GITHUB_SECRET` environment variable.

## Installing the app

In the app settings, click on "Install App" and give the app permissions to the repositories you want to use with Woodpecker.

## All GitHub Configuration Options

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

- `WOODPECKER_GITHUB` - Enables the GitHub driver (Default: `false`)

- `WOODPECKER_GITHUB_URL` - Configures the GitHub server address (Default: `https://github.com`)

- `WOODPECKER_GITHUB_CLIENT` - Configures the GitHub OAuth client id to authorize access (Default: empty)

- `WOODPECKER_GITHUB_CLIENT_FILE` - Read the value for `WOODPECKER_GITHUB_CLIENT` from the specified filepath (Default: empty)

- `WOODPECKER_GITHUB_SECRET` - Configures the GitHub OAuth client secret. This is used to authorize access. (Default: empty)

- `WOODPECKER_GITHUB_SECRET_FILE` - Read the value for `WOODPECKER_GITHUB_SECRET` from the specified filepath 9Default: empty)

  `WOODPECKER_GITHUB_MERGE_REF` - (Default: `true`)

- `WOODPECKER_GITHUB_SKIP_VERIFY` - Configure if SSL verification should be skipped (Default: `false`)
