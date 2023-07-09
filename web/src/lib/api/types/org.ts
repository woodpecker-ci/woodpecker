// A version control organization.
export type Org = {
  // The name of the organization.
  id: number;
  name: string;
  is_user: boolean;
};

export type OrgPermissions = {
  member: boolean;
  admin: boolean;
};
