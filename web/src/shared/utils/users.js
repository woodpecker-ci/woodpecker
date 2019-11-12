import { displayMessage } from "./message";

/**
* Generates a personal access token and stores the results in
* the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 */
export const generateToken = (tree, client) => {
	client
		.getToken()
		.then(token => {
			tree.set(["token"], token);
		})
		.catch(() => {
			displayMessage(tree, "Failed to retrieve your personal access token");
		});
};
