// The user account.
export type User = {
  id: number;
  // The unique identifier for the account.

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
};
