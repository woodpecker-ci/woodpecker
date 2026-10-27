// A version control organization.
export interface Org {
  // The name of the organization.
  id: number;
  forge_id: number;
  forge_remote_id: string;
  name: string;
  is_user: boolean;
  require_org_agents: boolean;
  // The forge remote id of another forge account that now uses the name of this org.
  conflict_forge_remote_id?: string;
}

export interface OrgPermissions {
  member: boolean;
  admin: boolean;
}
