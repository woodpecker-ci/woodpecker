export enum WebhookEvents {
  Push = 'push',
  Tag = 'tag',
  Release = 'release',
  PullRequest = 'pull_request',
  PullRequestClosed = 'pull_request_closed',
  Deploy = 'deployment',
  Cron = 'cron',
  Manual = 'manual',
}
