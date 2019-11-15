import { repositorySlug } from "./repository";

export function subscribeToLogs(tree, client, owner, repo, build, proc) {
  if (subscribeToLogs.ws) {
    subscribeToLogs.ws.close();
  }
  const slug = repositorySlug(owner, repo);
  const init = { data: [] };

  tree.set(["logs", "data", slug, build, proc.pid], init);

  subscribeToLogs.ws = client.stream(owner, repo, build, proc.ppid, item => {
    if (item.proc === proc.name) {
      tree.push(["logs", "data", slug, build, proc.pid, "data"], item);
    }
  });
}

export function fetchLogs(tree, client, owner, repo, build, proc) {
  const slug = repositorySlug(owner, repo);
  const init = {
    data: [],
    loading: true
  };

  tree.set(["logs", "data", slug, build, proc], init);

  client
    .getLogs(owner, repo, build, proc)
    .then(results => {
      tree.set(["logs", "data", slug, build, proc, "data"], results || []);
      tree.set(["logs", "data", slug, build, proc, "loading"], false);
      tree.set(["logs", "data", slug, build, proc, "eof"], true);
    })
    .catch(() => {
      tree.set(["logs", "data", slug, build, proc, "loading"], false);
      tree.set(["logs", "data", slug, build, proc, "eof"], true);
    });
}

/**
 * Toggles whether or not the browser should follow
 * the logs (ie scroll to bottom).
 *
 * @param {boolean} follow - Follow the logs.
 */
export const toggleLogs = (tree, follow) => {
  tree.set(["logs", "follow"], follow);
};
