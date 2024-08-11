package woodpecker

import "fmt"

const (
	pathGlobalVariables = "%s/api/variables"
	pathGlobalVariable  = "%s/api/variables/%s"
)

// GlobalVariable returns an global variable by name.
func (c *client) GlobalVariable(variable string) (*Variable, error) {
	out := new(Variable)
	uri := fmt.Sprintf(pathGlobalVariable, c.addr, variable)
	err := c.get(uri, out)
	return out, err
}

// GlobalVariableList returns a list of all global variables.
func (c *client) GlobalVariableList() ([]*Variable, error) {
	var out []*Variable
	uri := fmt.Sprintf(pathGlobalVariables, c.addr)
	err := c.get(uri, &out)
	return out, err
}

// GlobalVariableCreate creates a global variable.
func (c *client) GlobalVariableCreate(in *Variable) (*Variable, error) {
	out := new(Variable)
	uri := fmt.Sprintf(pathGlobalVariables, c.addr)
	err := c.post(uri, in, out)
	return out, err
}

// GlobalVariableUpdate updates a global variable.
func (c *client) GlobalVariableUpdate(in *Variable) (*Variable, error) {
	out := new(Variable)
	uri := fmt.Sprintf(pathGlobalVariable, c.addr, in.Name)
	err := c.patch(uri, in, out)
	return out, err
}

// GlobalVariableDelete deletes a global variable.
func (c *client) GlobalVariableDelete(variable string) error {
	uri := fmt.Sprintf(pathGlobalVariable, c.addr, variable)
	return c.delete(uri)
}
