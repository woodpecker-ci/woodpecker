export type ForgeType = 'github' | 'gitlab' | 'gitea' | 'bitbucket' | 'bitbucket-dc' | 'addon';

export type Forge = {
  id: number;
  type: ForgeType;
  url: string;
  client: string;
  clientSecret: string;
  skipVerify: boolean;
  oauthHost: string;
  additionalOptions: Record<string, string>;
};
