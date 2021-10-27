/**
 * Get the event feed and store the results in the
 * state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 */
export const fetchFeed = (tree, client) => {
  client
    .getBuildFeed({ latest: true })
    .then(results => {
      let list = {};
      let sorted = results.sort(compareFeedItem);
      sorted.map(repo => {
        list[repo.full_name] = repo;
      });
      if (sorted && sorted.length > 0) {
        tree.set(["feed", "latest"], sorted[0]);
      }
      tree.set(["feed", "loaded"], true);
      tree.set(["feed", "data"], list);
    })
    .catch(error => {
      tree.set(["feed", "loaded"], true);
      tree.set(["feed", "error"], error);
    });
};

/**
 * Ensures the fetchFeed function is invoked exactly once.
 * TODO replace this with a decorator
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 */
export function fetchFeedOnce(tree, client) {
  if (fetchFeedOnce.fired) {
    return;
  }
  fetchFeedOnce.fired = true;
  return fetchFeed(tree, client);
}

/**
 * Subscribes to the server-side event feed and synchronizes
 * event data with the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 */
export const subscribeToFeed = (tree, client) => {
  return client.on(data => {
    const { repo, build } = data;

    if (tree.exists("feed", "data", repo.full_name)) {
      const cursor = tree.select(["feed", "data", repo.full_name]);
      cursor.merge(build);
    }

    if (tree.exists("builds", "data", repo.full_name)) {
      tree.set(["builds", "data", repo.full_name, build.number], build);
    }
  });
};

/**
 * Ensures the subscribeToFeed function is invoked exactly once.
 * TODO replace this with a decorator
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 */
export function subscribeToFeedOnce(tree, client) {
  if (subscribeToFeedOnce.fired) {
    return;
  }
  subscribeToFeedOnce.fired = true;
  return subscribeToFeed(tree, client);
}

/**
 * Compare two feed items by name.
 * @param {Object} a - A feed item.
 * @param {Object} b - A feed item.
 * @returns {number}
 */
export const compareFeedItem = (a, b) => {
  return (
    (b.started_at || b.created_at || -1) - (a.started_at || a.created_at || -1)
  );
};
