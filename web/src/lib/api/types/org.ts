// A version control organization.
export type Org = {
  // The name of the organization.
  name: string;
};

export type OrgPermissions = {
  member: boolean;
  admin: boolean;
};
