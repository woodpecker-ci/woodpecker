package woodpecker

import (
	"fmt"
	"net/url"
)

const (
	pathSelf  = "%s/api/user"
	pathRepos = "%s/api/user/repos"
	pathUsers = "%s/api/users"
	pathUser  = "%s/api/users/%s"
)

type RepoListOptions struct {
	All bool // query all repos, including inactive ones
}

// QueryEncode returns the URL query parameters for the RepoListOptions.
func (opt *RepoListOptions) QueryEncode() string {
	query := make(url.Values)
	if opt.All {
		query.Add("all", "true")
	}
	return query.Encode()
}

// Self returns the currently authenticated user.
func (c *client) Self() (*User, error) {
	out := new(User)
	uri := fmt.Sprintf(pathSelf, c.addr)
	err := c.get(uri, out)
	return out, err
}

// User returns a user by login.
func (c *client) User(login string) (*User, error) {
	out := new(User)
	uri := fmt.Sprintf(pathUser, c.addr, login)
	err := c.get(uri, out)
	return out, err
}

// UserList returns a list of all registered users.
func (c *client) UserList() ([]*User, error) {
	var out []*User
	uri := fmt.Sprintf(pathUsers, c.addr)
	err := c.get(uri, &out)
	return out, err
}

// UserPost creates a new user account.
func (c *client) UserPost(in *User) (*User, error) {
	out := new(User)
	uri := fmt.Sprintf(pathUsers, c.addr)
	err := c.post(uri, in, out)
	return out, err
}

// UserPatch updates a user account.
func (c *client) UserPatch(in *User) (*User, error) {
	out := new(User)
	uri := fmt.Sprintf(pathUser, c.addr, in.Login)
	err := c.patch(uri, in, out)
	return out, err
}

// UserDel deletes a user account.
func (c *client) UserDel(login string) error {
	uri := fmt.Sprintf(pathUser, c.addr, login)
	err := c.delete(uri)
	return err
}

// RepoList returns a list of all repositories to which
// the user has explicit access in the host system.
func (c *client) RepoList(opt RepoListOptions) ([]*Repo, error) {
	var out []*Repo

	uri, _ := url.Parse(fmt.Sprintf(pathRepos, c.addr))
	uri.RawQuery = opt.QueryEncode()

	err := c.get(uri.String(), &out)
	return out, err
}
