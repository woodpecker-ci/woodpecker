import { displayMessage } from "./message";
import { fetchFeed } from "shared/utils/feed";

/**
 * Get the named repository and store the results in
 * the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 */
export const fetchRepository = (tree, client, owner, name) => {
  tree.unset(["repo", "error"]);
  tree.unset(["repo", "loaded"]);

  client
    .getRepo(owner, name)
    .then(repo => {
      tree.set(["repos", "data", repo.full_name], repo);
      tree.set(["repo", "loaded"], true);
    })
    .catch(error => {
      tree.set(["repo", "error"], error);
      tree.set(["repo", "loaded"], true);
    });
};

/**
 * Get the repository list for the current user and
 * store the results in the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 */
export const fetchRepostoryList = (tree, client) => {
  tree.unset(["repos", "loaded"]);
  tree.unset(["repos", "error"]);

  client
    .getRepoList({ all: true })
    .then(results => {
      let list = {};
      results.map(repo => {
        list[repo.full_name] = repo;
      });

      const path = ["repos", "data"];
      if (tree.exists(path)) {
        tree.deepMerge(path, list);
      } else {
        tree.set(path, list);
      }

      tree.set(["repos", "loaded"], true);
    })
    .catch(error => {
      tree.set(["repos", "loaded"], true);
      tree.set(["repos", "error"], error);
    });
};

/**
 * Synchronize the repository list for the current user
 * and merge the results into the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 */
export const syncRepostoryList = (tree, client) => {
  tree.unset(["repos", "loaded"]);
  tree.unset(["repos", "error"]);

  client
    .getRepoList({ all: true, flush: true })
    .then(results => {
      let list = {};
      results.map(repo => {
        list[repo.full_name] = repo;
      });

      const path = ["repos", "data"];
      if (tree.exists(path)) {
        tree.deepMerge(path, list);
      } else {
        tree.set(path, list);
      }

      displayMessage(tree, "Successfully synchronized your repository list");
      tree.set(["repos", "loaded"], true);
    })
    .catch(error => {
      displayMessage(tree, "Failed to synchronize your repository list");
      tree.set(["repos", "loaded"], true);
      tree.set(["repos", "error"], error);
    });
};

/**
 * Update the repository and if successful update the
 * repository information into the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 * @param {Object} data - The repository updates.
 */
export const updateRepository = (tree, client, owner, name, data) => {
  client
    .updateRepo(owner, name, data)
    .then(repo => {
      tree.set(["repos", "data", repo.full_name], repo);
      displayMessage(tree, "Successfully updated the repository settings");
    })
    .catch(() => {
      displayMessage(tree, "Failed to update the repository settings");
    });
};

/**
 * Enables the repository and if successful update the
 * repository active status in the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 */
export const enableRepository = (tree, client, owner, name) => {
  client
    .activateRepo(owner, name)
    .then(result => {
      displayMessage(tree, "Successfully activated your repository");
      tree.set(["repos", "data", result.full_name, "active"], true);
      fetchFeed(tree, client);
    })
    .catch(() => {
      displayMessage(tree, "Failed to activate your repository");
    });
};

/**
 * Disables the repository and if successful update the
 * repository active status in the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 */
export const disableRepository = (tree, client, owner, name) => {
  client
    .deleteRepo(owner, name)
    .then(result => {
      displayMessage(tree, "Successfully disabled your repository");
      tree.set(["repos", "data", result.full_name, "active"], false);
      fetchFeed(tree, client);
    })
    .catch(() => {
      displayMessage(tree, "Failed to disabled your repository");
    });
};

/**
 * Compare two repositories by name.
 *
 * @param {Object} a - A repository.
 * @param {Object} b - A repository.
 * @returns {number}
 */
export const compareRepository = (a, b) => {
  if (a.full_name < b.full_name) return -1;
  if (a.full_name > b.full_name) return 1;
  return 0;
};

/**
 * Returns the repository slug.
 *
 * @param {string} owner - The repository owner.
 * @param {string} name - The process name.
 */
export const repositorySlug = (owner, name) => {
  return `${owner}/${name}`;
};
