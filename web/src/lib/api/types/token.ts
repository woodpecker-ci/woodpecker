/* eslint-disable no-unused-vars */
export enum TokenType {
  Badge = 'badge',
}
/* eslint-enable */

// A secret bound to a repo, granting access to a single feature of it.
export interface Token {
  id: number;
  repo_id: number;
  type: TokenType;
  value: string;
  created: number;
}
