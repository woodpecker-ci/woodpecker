/* eslint-disable no-unused-vars */
export enum WebhookEvents {
  Push = 'push',
  Tag = 'tag',
  Release = 'release',
  PullRequest = 'pull_request',
  PullRequestClosed = 'pull_request_closed',
  PullRequestMetadata = 'pull_request_metadata',
  Deploy = 'deployment',
  Cron = 'cron',
  Manual = 'manual',
}
/* eslint-enable */
