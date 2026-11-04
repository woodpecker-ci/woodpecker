import axios, { AxiosError } from 'axios';

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

// getLatestReleaseTag returns the tag of the latest release of the repository a docs file belongs to,
// without a leading "v" (v1.2.3 becomes 1.2.3), or undefined if there is none or it can't be fetched.
export async function getLatestReleaseTag(docsUrl: string): Promise<string | undefined> {
  const apiUrl = releaseApiUrl(docsUrl);
  if (!apiUrl) {
    return undefined;
  }

  try {
    const response = await axios<{ tag_name?: string }>(apiUrl);
    return response.data.tag_name?.replace(/^v(?=\d)/, '') || undefined;
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
