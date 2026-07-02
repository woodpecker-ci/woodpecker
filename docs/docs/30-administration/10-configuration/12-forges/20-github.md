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

## App Settings

- Name: An arbitrary name for your App
- Homepage URL: The URL of your Woodpecker instance
- Callback URL: `https://<your-woodpecker-instance>/authorize`
- (optional) Upload the Woodpecker Logo: <https://avatars.githubusercontent.com/u/84780935?s=200&v=4>

## Client Secret Creation

After your App has been created, you can generate a client secret.
Use this one for the `WOODPECKER_GITHUB_SECRET` environment variable.

## GitHub App

In addition to the OAuth login, Woodpecker can authenticate as a [GitHub App](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps) using installation access tokens for server-side API calls (commit statuses, pipeline configuration fetching, changed-file lookups) and for cloning repositories.
This is recommended for larger organizations:

- API calls are counted against the [rate limit of the app installation](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api#primary-rate-limit-for-github-app-installations) (5,000 requests/hour and more, scaling with the number of repositories) instead of the personal rate limit of the user who enabled the repository.
- Pipelines keep working even if the user who enabled the repository loses access or leaves the organization.
- Commit statuses are posted by the app's bot account instead of a personal user account.

To use it, create a GitHub App (Settings -> Developer Settings -> GitHub Apps -> New GitHub App) and install it on your organization or the repositories you want to build.
Disable the app's own webhook (uncheck "Active" under "Webhook") — Woodpecker keeps creating a webhook on each enabled repository itself.
Grant the following repository permissions:

- `Contents`: `Read-only` (clone repositories and read pipeline configuration)
- `Commit statuses`: `Read and write` (report pipeline status)
- `Pull requests`: `Read-only` (list changed files of pull requests)
- `Deployments`: `Read and write` (only needed to report the status of deployment pipelines)

Then configure the app credentials in addition to the OAuth settings. An OAuth client is still required for user login: keep using your OAuth2 app for `WOODPECKER_GITHUB_CLIENT`/`WOODPECKER_GITHUB_SECRET`. (The GitHub App's own OAuth credentials work for login too, but user tokens issued by a GitHub App only grant access to repositories the app is installed on, which limits repository listing and the user-token fallback described below.)

```ini
WOODPECKER_GITHUB_APP_ID=YOUR_GITHUB_APP_OR_CLIENT_ID
WOODPECKER_GITHUB_APP_PRIVATE_KEY_FILE=/etc/woodpecker/github-app-private-key.pem
```

For repositories the app is not installed on, Woodpecker falls back to the OAuth token of the user who enabled the repository.

The private key is write-only: the API and the admin UI never return it, and leaving the field empty when updating a forge keeps the stored key.
After saving, the "Test GitHub App" button in the admin UI (or `GET /api/forges/{id}/app-health`) verifies that the credentials work.

:::note
Installation access tokens expire after one hour and are scoped to the repositories of the app installation.
Clone credentials of pipelines that stay queued for a very long time may therefore expire before the clone step runs, and clone access to private repositories outside the installation (e.g. git submodules in another organization) requires falling back to user OAuth tokens by not installing the app on the affected repositories.
:::

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

---

### GITHUB_APP_ID

- Name: `WOODPECKER_GITHUB_APP_ID`
- Default: none

Configures the app id (or client id) of a GitHub App. When set together with the private key, Woodpecker uses installation access tokens of the app for server-side API calls and for cloning repositories the app is installed on.

---

### GITHUB_APP_PRIVATE_KEY

- Name: `WOODPECKER_GITHUB_APP_PRIVATE_KEY`
- Default: none

Configures the private key of the GitHub App, either as plain PEM or base64-encoded PEM.

---

### GITHUB_APP_PRIVATE_KEY_FILE

- Name: `WOODPECKER_GITHUB_APP_PRIVATE_KEY_FILE`
- Default: none

Read the value for `WOODPECKER_GITHUB_APP_PRIVATE_KEY` from the specified filepath, e.g. the `.pem` file downloaded from GitHub.
