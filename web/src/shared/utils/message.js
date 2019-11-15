/**
 * Displays the global message.
 *
 * @param {Object} tree - The drone state tree.
 * @param {string} message - The message text.
 */
export const displayMessage = (tree, message) => {
  tree.set(["message", "text"], message);

  setTimeout(() => {
    hideMessage(tree);
  }, 5000);
};

/**
 * Hide the global message.
 *
 * @param {Object} tree - The drone state tree.
 */
export const hideMessage = tree => {
  tree.unset(["message", "text"]);
};
