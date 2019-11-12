import { displayMessage } from "./message";
import { repositorySlug } from "./repository";

/**
 * Get the registry list for the named repository and
 * store the results in the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 */
export const fetchRegistryList = (tree, client, owner, name) => {
	const slug = repositorySlug(owner, name);

	tree.unset(["registry", "loaded"]);
	tree.unset(["registry", "error"]);

	client.getRegistryList(owner, name).then(results => {
		let list = {};
		results.map(registry => {
			list[registry.address] = registry;
		});
		tree.set(["registry", "data", slug], list);
		tree.set(["registry", "loaded"], true);
	});
};

/**
 * Create the registry credentials for the named repository
 * and if successful, store the result in the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 * @param {Object} registry - The registry hostname.
 */
export const createRegistry = (tree, client, owner, name, registry) => {
	const slug = repositorySlug(owner, name);

	client
		.createRegistry(owner, name, registry)
		.then(result => {
			tree.set(["registry", "data", slug, registry.address], result);
			displayMessage(tree, "Successfully stored the registry credentials");
		})
		.catch(() => {
			displayMessage(tree, "Failed to store the registry credentials");
		});
};

/**
 * Delete the registry credentials for the named repository
 * and if successful, remove from the state tree.
 *
 * @param {Object} tree - The drone state tree.
 * @param {Object} client - The drone client.
 * @param {string} owner - The repository owner.
 * @param {string} name - The repository name.
 * @param {Object} registry - The registry hostname.
 */
export const deleteRegistry = (tree, client, owner, name, registry) => {
	const slug = repositorySlug(owner, name);

	client
		.deleteRegistry(owner, name, registry)
		.then(result => {
			tree.unset(["registry", "data", slug, registry]);
			displayMessage(tree, "Successfully deleted the registry credentials");
		})
		.catch(() => {
			displayMessage(tree, "Failed to delete the registry credentials");
		});
};
