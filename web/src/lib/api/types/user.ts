// The user account.
export interface User {
  id: number;
  // The unique identifier for the account.

  forge_id: number;
  // The unique identifier of the forge the account belongs to.

  forge_remote_id: string;
  // The unique identifier of user at the remote forge.

  login: string;
  // The login name for the account.

  email: string;
  // The email address for the account.

  avatar_url: string;
  // The url for the avatar image.

  admin: boolean;
  // Whether the account has administrative privileges.

  active: boolean;
  // Whether the account is currently active.

  org_id: number;
  // The ID of the org assigned to the user.
}
