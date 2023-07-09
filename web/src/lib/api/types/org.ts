// A version control organization.
export type Org = {
  // The name of the organization.
  id: number;
  name: string;
  type: 'user' | 'team';
};

export type OrgPermissions = {
  member: boolean;
  admin: boolean;
};
