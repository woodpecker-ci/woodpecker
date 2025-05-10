/* eslint-disable no-unused-vars */
export enum WebhookEvents {
  Push = 'push',
  Tag = 'tag',
  Release = 'release',
  PullRequest = 'pull_request',
  PullRequestClosed = 'pull_request_closed',
  PullRequestEdited = 'pull_request_edited',
  Deploy = 'deployment',
  Cron = 'cron',
  Manual = 'manual',
}
/* eslint-enable */
