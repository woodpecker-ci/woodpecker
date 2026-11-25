---
toc_max_heading_level: 2
---

# GitHub

Woodpecker comes with built-in support for GitHub and GitHub Enterprise.
To use Woodpecker with GitHub the following environment variables should be set for the server component:

```ini
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

---

### GITHUB

- Name: `WOODPECKER_GITHUB`
- Default: `false`

Enables the GitHub driver.

---

### GITHUB_URL

- Name: `WOODPECKER_GITHUB_URL`
- Default: `https://github.com`

Configures the GitHub server address.

---

### GITHUB_CLIENT

- Name: `WOODPECKER_GITHUB_CLIENT`
- Default: none

Configures the GitHub OAuth client id to authorize access.

---

### GITHUB_CLIENT_FILE

- Name: `WOODPECKER_GITHUB_CLIENT_FILE`
- Default: none

Read the value for `WOODPECKER_GITHUB_CLIENT` from the specified filepath.

---

### GITHUB_SECRET

- Name: `WOODPECKER_GITHUB_SECRET`
- Default: none

Configures the GitHub OAuth client secret. This is used to authorize access.

---

### GITHUB_SECRET_FILE

- Name: `WOODPECKER_GITHUB_SECRET_FILE`
- Default: none

Read the value for `WOODPECKER_GITHUB_SECRET` from the specified filepath.

---

### GITHUB_MERGE_REF

- Name: `WOODPECKER_GITHUB_MERGE_REF`
- Default: `true`

---

### GITHUB_SKIP_VERIFY

- Name: `WOODPECKER_GITHUB_SKIP_VERIFY`
- Default: `false`

Configure if SSL verification should be skipped.

---

### GITHUB_PUBLIC_ONLY

- Name: `WOODPECKER_GITHUB_PUBLIC_ONLY`
- Default: `false`

Configures the GitHub OAuth client to only obtain a token that can manage public repositories.
