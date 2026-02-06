export type ForgeType = 'github' | 'gitlab' | 'gitea' | 'bitbucket' | 'bitbucket-dc' | 'addon' | 'forgejo';

export interface Forge {
  id: number;
  type: ForgeType;
  url: string;
  client?: string;
  oauth_client_secret?: string;
  skip_verify?: boolean;
  oauth_host?: string;
  additional_options?: Record<string, unknown>;
}
