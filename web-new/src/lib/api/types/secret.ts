export enum SecretEvents {
  Push = 'push',
  Tag = 'tag',
  PullRequest = 'pull-request',
  Deploy = 'deploy',
}

export type Secret = {
  id: string;
  name: string;
  value: string;
  event: SecretEvents[];
  image: string[];
};
