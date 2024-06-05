package woodpecker

import "fmt"

const (
	pathGlobalSecrets = "%s/api/secrets"
	pathGlobalSecret  = "%s/api/secrets/%s"
)

// GlobalSecret returns an global secret by name.
func (c *client) GlobalSecret(secret string) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathGlobalSecret, c.addr, secret)
	err := c.get(uri, out)
	return out, err
}

// GlobalSecretList returns a list of all global secrets.
func (c *client) GlobalSecretList() ([]*Secret, error) {
	var out []*Secret
	uri := fmt.Sprintf(pathGlobalSecrets, c.addr)
	err := c.get(uri, &out)
	return out, err
}

// GlobalSecretCreate creates a global secret.
func (c *client) GlobalSecretCreate(in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathGlobalSecrets, c.addr)
	err := c.post(uri, in, out)
	return out, err
}

// GlobalSecretUpdate updates a global secret.
func (c *client) GlobalSecretUpdate(in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathGlobalSecret, c.addr, in.Name)
	err := c.patch(uri, in, out)
	return out, err
}

// GlobalSecretDelete deletes a global secret.
func (c *client) GlobalSecretDelete(secret string) error {
	uri := fmt.Sprintf(pathGlobalSecret, c.addr, secret)
	return c.delete(uri)
}
