package woodpecker

import "fmt"

const (
	pathOrg           = "%s/api/orgs/%d"
	pathOrgLookup     = "%s/api/orgs/lookup/%s"
	pathOrgSecrets    = "%s/api/orgs/%d/secrets"
	pathOrgSecret     = "%s/api/orgs/%d/secrets/%s"
	pathOrgRegistries = "%s/api/orgs/%d/registries"
	pathOrgRegistry   = "%s/api/orgs/%d/registries/%s"
)

// Org returns an organization by id.
func (c *client) Org(orgID int64) (*Org, error) {
	out := new(Org)
	uri := fmt.Sprintf(pathOrg, c.addr, orgID)
	err := c.get(uri, out)
	return out, err
}

// OrgLookup returns a organization by its name.
func (c *client) OrgLookup(name string) (*Org, error) {
	out := new(Org)
	uri := fmt.Sprintf(pathOrgLookup, c.addr, name)
	err := c.get(uri, out)
	return out, err
}

// OrgSecret returns an organization secret by name.
func (c *client) OrgSecret(orgID int64, secret string) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathOrgSecret, c.addr, orgID, secret)
	err := c.get(uri, out)
	return out, err
}

// OrgSecretList returns a list of all organization secrets.
func (c *client) OrgSecretList(orgID int64) ([]*Secret, error) {
	var out []*Secret
	uri := fmt.Sprintf(pathOrgSecrets, c.addr, orgID)
	err := c.get(uri, &out)
	return out, err
}

// OrgSecretCreate creates an organization secret.
func (c *client) OrgSecretCreate(orgID int64, in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathOrgSecrets, c.addr, orgID)
	err := c.post(uri, in, out)
	return out, err
}

// OrgSecretUpdate updates an organization secret.
func (c *client) OrgSecretUpdate(orgID int64, in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathOrgSecret, c.addr, orgID, in.Name)
	err := c.patch(uri, in, out)
	return out, err
}

// OrgSecretDelete deletes an organization secret.
func (c *client) OrgSecretDelete(orgID int64, secret string) error {
	uri := fmt.Sprintf(pathOrgSecret, c.addr, orgID, secret)
	return c.delete(uri)
}

// OrgRegistry returns an organization registry by address.
func (c *client) OrgRegistry(orgID int64, registry string) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathOrgRegistry, c.addr, orgID, registry)
	err := c.get(uri, out)
	return out, err
}

// OrgRegistryList returns a list of all organization registries.
func (c *client) OrgRegistryList(orgID int64) ([]*Registry, error) {
	var out []*Registry
	uri := fmt.Sprintf(pathOrgRegistries, c.addr, orgID)
	err := c.get(uri, &out)
	return out, err
}

// OrgRegistryCreate creates an organization registry.
func (c *client) OrgRegistryCreate(orgID int64, in *Registry) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathOrgRegistries, c.addr, orgID)
	err := c.post(uri, in, out)
	return out, err
}

// OrgRegistryUpdate updates an organization registry.
func (c *client) OrgRegistryUpdate(orgID int64, in *Registry) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathOrgRegistry, c.addr, orgID, in.Address)
	err := c.patch(uri, in, out)
	return out, err
}

// OrgRegistryDelete deletes an organization registry.
func (c *client) OrgRegistryDelete(orgID int64, registry string) error {
	uri := fmt.Sprintf(pathOrgRegistry, c.addr, orgID, registry)
	return c.delete(uri)
}
