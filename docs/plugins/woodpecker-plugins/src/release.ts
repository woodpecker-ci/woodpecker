import axios, { AxiosError } from 'axios';

import { WoodpeckerPluginRelease } from './types';

// apiTokens lists the optional read tokens per API host, to avoid rate limits of anonymous requests.
const apiTokens: Record<string, { env: string; scheme: string }> = {
  'api.github.com': { env: 'GITHUB_TOKEN', scheme: 'Bearer' },
  'codeberg.org': { env: 'CODEBERG_TOKEN', scheme: 'token' },
};

// authHeaders returns the authorization header for an API url, if a token for its host is set.
function authHeaders(apiUrl: string): Record<string, string> {
  const token = apiTokens[new URL(apiUrl).hostname];
  const value = token && process.env[token.env];
  return value ? { Authorization: `${token.scheme} ${value}` } : {};
}

// releaseApiUrl returns the API url of the latest release of the repository a docs file belongs to,
// or undefined if the forge is not known.
export function releaseApiUrl(docsUrl: string): string | undefined {
  let url: URL;
  try {
    url = new URL(docsUrl);
  } catch {
    return undefined;
  }

  const [owner, repo, ...rest] = url.pathname.split('/').filter((part) => part !== '');
  if (!owner || !repo) {
    return undefined;
  }

  // GitHub: raw.githubusercontent.com/OWNER/REPO/... or github.com/OWNER/REPO/raw/...
  if (url.hostname === 'raw.githubusercontent.com' || (url.hostname === 'github.com' && rest[0] === 'raw')) {
    return `https://api.github.com/repos/${owner}/${repo}/releases/latest`;
  }

  // Gitea, Forgejo and Codeberg: HOST/OWNER/REPO/raw/(branch|tag|commit)/...
  if (rest[0] === 'raw' && ['branch', 'tag', 'commit'].includes(rest[1])) {
    return `${url.origin}/api/v1/repos/${owner}/${repo}/releases/latest`;
  }

  return undefined;
}

// getLatestRelease returns the latest release of the repository a docs file belongs to, with the tag
// without a leading "v" (v1.2.3 becomes 1.2.3), or undefined if there is none or it can't be fetched.
export async function getLatestRelease(docsUrl: string): Promise<WoodpeckerPluginRelease | undefined> {
  const apiUrl = releaseApiUrl(docsUrl);
  if (!apiUrl) {
    return undefined;
  }

  try {
    const response = await axios<{ tag_name?: string; html_url?: string }>(apiUrl, { headers: authHeaders(apiUrl) });
    const tag = response.data.tag_name?.replace(/^v(?=\d)/, '');
    return tag ? { tag, url: response.data.html_url } : undefined;
  } catch (e) {
    const axiosError = e as AxiosError;
    // no release published
    if (axiosError.response?.status === 404) {
      return undefined;
    }
    console.error(
      "Can't fetch latest release",
      apiUrl,
      axiosError.message,
      axiosError.response?.status,
      axiosError.response?.statusText,
    );
    return undefined;
  }
}
