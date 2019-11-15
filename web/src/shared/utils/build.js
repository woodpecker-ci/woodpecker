import { repositorySlug } from "./repository";
import { displayMessage } from "./message";
import { STATUS_PENDING, STATUS_RUNNING } from "shared/constants/status";

/**
 * Gets the build for the named repository and stores
 * the results in the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 * @param {number|string} number - The build number.
 */
export const fetchBuild = (tree, client, owner, name, number) => {
  const slug = repositorySlug(owner, name);

  tree.unset(["builds", "loaded"]);
  client
    .getBuild(owner, name, number)
    .then(build => {
      const path = ["builds", "data", slug, build.number];

      if (tree.exists(path)) {
        tree.deepMerge(path, build);
      } else {
        tree.set(path, build);
      }

      tree.set(["builds", "loaded"], true);
    })
    .catch(error => {
      tree.set(["builds", "loaded"], true);
      tree.set(["builds", "error"], error);
    });
};

/**
 * Gets the build list for the named repository and
 * stores the results in the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 */
export const fetchBuildList = (tree, client, owner, name, page = 1) => {
  const slug = repositorySlug(owner, name);

  tree.unset(["builds", "loaded"]);
  tree.unset(["builds", "error"]);

  client
    .getBuildList(owner, name, { page: page })
    .then(results => {
      let list = {};
      results.map(build => {
        list[build.number] = build;
      });

      const path = ["builds", "data", slug];
      if (tree.exists(path)) {
        tree.deepMerge(path, list);
      } else {
        tree.set(path, list);
      }

      tree.unset(["builds", "error"]);
      tree.set(["builds", "loaded"], true);
    })
    .catch(error => {
      tree.set(["builds", "error"], error);
      tree.set(["builds", "loaded"], true);
    });
};

/**
 * Cancels the build.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 * @param {number} build - The build number.
 * @param {number} proc - The process number.
 */
export const cancelBuild = (tree, client, owner, repo, build, proc) => {
  client
    .cancelBuild(owner, repo, build, proc)
    .then(result => {
      displayMessage(tree, "Successfully cancelled your build");
    })
    .catch(() => {
      displayMessage(tree, "Failed to cancel your build");
    });
};

/**
 * Restarts the build.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 * @param {number} build - The build number.
 */
export const restartBuild = (tree, client, owner, repo, build) => {
  client
    .restartBuild(owner, repo, build, { fork: true })
    .then(result => {
      displayMessage(tree, "Successfully restarted your build");
    })
    .catch(() => {
      displayMessage(tree, "Failed to restart your build");
    });
};

/**
 * Approves the blocked build.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 * @param {number} build - The build number.
 */
export const approveBuild = (tree, client, owner, repo, build) => {
  client
    .approveBuild(owner, repo, build)
    .then(result => {
      displayMessage(tree, "Successfully processed your approval decision");
    })
    .catch(() => {
      displayMessage(tree, "Failed to process your approval decision");
    });
};

/**
 * Declines the blocked build.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 * @param {number} build - The build number.
 */
export const declineBuild = (tree, client, owner, repo, build) => {
  client
    .declineBuild(owner, repo, build)
    .then(result => {
      displayMessage(tree, "Successfully processed your decline decision");
    })
    .catch(() => {
      displayMessage(tree, "Failed to process your decline decision");
    });
};

/**
 * Compare two builds by number.
 *
 * @param {Object} a - A build.
 * @param {Object} b - A build.
 * @returns {number}
 */
export const compareBuild = (a, b) => {
  return b.number - a.number;
};

/**
 * Returns true if the build is in a penidng or running state.
 *
 * @param {Object} build - The build object.
 * @returns {boolean}
 */
export const assertBuildFinished = build => {
  return build.status !== STATUS_RUNNING && build.status !== STATUS_PENDING;
};

/**
 * Returns true if the build is a matrix.
 *
 * @param {Object} build - The build object.
 * @returns {boolean}
 */
export const assertBuildMatrix = build => {
  return build && build.procs && build.procs.length > 1;
};
