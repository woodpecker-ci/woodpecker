export interface Registry {
  id: string;
  repo_id: number;
  org_id: number;
  address: string;
  username: string;
  password: string;
  readonly: boolean;
}
