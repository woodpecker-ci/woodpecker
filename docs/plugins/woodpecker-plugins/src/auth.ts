// hostTokens lists the optional read tokens per forge host, to avoid rate limits of anonymous requests.
const hostTokens: Record<string, { env: string; scheme: string }> = {
  'api.github.com': { env: 'GITHUB_TOKEN', scheme: 'Bearer' },
  'raw.githubusercontent.com': { env: 'GITHUB_TOKEN', scheme: 'Bearer' },
  'codeberg.org': { env: 'CODEBERG_TOKEN', scheme: 'token' },
};

// authHeaders returns the authorization header for a url, if a token for its host is set.
export function authHeaders(url: string): Record<string, string> {
  const token = hostTokens[new URL(url).hostname];
  const value = token && process.env[token.env];
  return value ? { Authorization: `${token.scheme} ${value}` } : {};
}
